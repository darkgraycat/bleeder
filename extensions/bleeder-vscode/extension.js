// @ts-check
const vscode = require('vscode');
const net = require('node:net');;
const chp = require('node:child_process');
const { regExs, getSequenceRaw, getSequenceDetails, parseBleedData } = require('./helpers');

/** @type {{
 * data?: ReturnType<typeof parseBleedData> // TODO remove
 * client: net.Socket
 * procs: Record<string, chp.ChildProcess>
 * stateUpdated: vscode.EventEmitter
 * }} 
 */
const state = {
  client: null,
  procs: {},
  stateUpdated: new vscode.EventEmitter()
};

/** @param {vscode.ExtensionContext} context */
function activate(context) {
  const activeDoc = getActiveDocument();
  if (activeDoc)
    state.data =
      parseBleedData(activeDoc.getText());

  context.subscriptions.push(
    // Events
    vscode.workspace.onDidSaveTextDocument(document => {
      state.data = parseBleedData(document.getText());
      console.log(state.data);
    }),
    // CodeLens
    vscode.languages.registerCodeLensProvider('toml', {
      onDidChangeCodeLenses: state.stateUpdated.event,
      provideCodeLenses(document) {
        /** @type {vscode.CodeLens[]} */
        const out = [];
        const text = document.getText();
        let match;
        while (match = regExs.seqDef.exec(text)) {
          const p = document.positionAt(match.index);
          const [, seqType, seqName] = match;
          const range = new vscode.Range(p.line, 0, p.line, 0);
          const isPlaying = state.procs[seqName];

          out.push(
            new vscode.CodeLens(range, {
              title: `${isPlaying ? '⏹' : '▶'} ${seqName}`,
              command: `bleeder.${isPlaying ? 'stop' : 'play'}`,
              arguments: [seqName],
            }),
          );
        }
        return out;
      }
    }),
    // InlayHints
    vscode.languages.registerInlayHintsProvider('toml', {
      provideInlayHints(document) {
        /** @type {vscode.InlayHint[]} */
        const out = [];
        const text = document.getText();
        const regex = /@(\w+):(\w+)/g; // @seq:args // TODO: remove
        let match;
        while (match = regex.exec(text)) {
          const p = document.positionAt(match.index);
          out.push({
            position: document.lineAt(p.line).range.end,
            label: ' [4 beats, tune: d1]',
            kind: vscode.InlayHintKind.Type,
            paddingLeft: true
          });
        }
        return out;
      }
    }),
    // Commands
    vscode.commands.registerCommand('bleeder.play', (seqName = 'main') => {
      const activeDoc = getActiveDocument();
      if (!activeDoc) return;

      vscode.window.setStatusBarMessage(`Playing: ${seqName}`, 5000);
      const filePath = activeDoc.uri.fsPath;
      const config = vscode.workspace.getConfiguration('bleeder');
      const binPath = config.get('path');
      const player = config.get('player');

      const cmd = `${binPath} play -seq ${seqName} ${filePath} | ${player}`;
      const proc = chp.spawn('sh', ['-c', cmd], { detached: true });
      state.procs[seqName] = proc;
      state.stateUpdated.fire();

      let stderr = '';
      proc.stderr.on('data', (data) => stderr += data.toString());
      proc.on('error', (err) => vscode.window.showErrorMessage(`Error: ${err.message}`));
      proc.on('close', (code) => {
        state.procs[seqName] = null;
        state.stateUpdated.fire();
        if (code !== 0) {
          // const last = stderr.split('\n').at(-2);
          // vscode.window.showErrorMessage(`Play failed: ${last}`);
          // TODO proper error handling
        }
      });
    }),
    vscode.commands.registerCommand('bleeder.stop', (seqName = 'main') => {
      const proc = state.procs[seqName];
      if (proc)
        process.kill(-proc.pid, 'SIGTERM');
    }),
    // Misc
    state.stateUpdated,
  );
}

function deactivate() { }

function getActiveDocument() {
  const doc = vscode.window.activeTextEditor?.document;
  if (!doc) vscode.window.showWarningMessage('No file open');
  return doc || null;
}

module.exports = { activate, deactivate };
