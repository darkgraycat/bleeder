// @ts-check
const vscode = require('vscode');
const net = require('net');

/** @param {vscode.ExtensionContext} context */
function activate(context) {
  /** @type {net.Socket} */
  let client = null;

  // CodeLens
  vscode.languages.registerCodeLensProvider('toml', {
    provideCodeLenses(document) {
      /** @type {vscode.CodeLens[]} */
      const out = [];
      const text = document.getText();
      const regex = /^\[(lane|riff)\.(\w+)\]/gm; // [lane.xxx] or [riff.xxx]
      let match;
      while (match = regex.exec(text)) {
        const p = document.positionAt(match.index);
        const [, seqType, seqName] = match;
        out.push(
          new vscode.CodeLens(
            new vscode.Range(p.line, 0, p.line, 0),
            {
              title: `Play ${seqName}`,
              command: 'bleeder.play',
              arguments: [seqName]
            })
        );
      }
      return out;
    }
  });

  // InlayHints
  vscode.languages.registerInlayHintsProvider('toml', {
    provideInlayHints(document) {
      /** @type {vscode.InlayHint[]} */
      const out = [];
      const text = document.getText();
      const regex = /@(\w+):(\w+)/g; // @seq:args
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
  });

  // Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('bleeder.play', (seqName, seqType) => {
      const msg = `play:${seqType}:${seqName}\n`;

      if (!client) {
        client = net.createConnection({ host: 'localhost', port: 9999 }, () => {
          client.write(msg);
          vscode.window.showInformationMessage(`Playing ${seqType}.${seqName}`);
        });

        client.on('error', () => {
          vscode.window.showErrorMessage('Bleeder server not running');
          client = null;
        });

        client.on('close', () => {
          client = null;
        });
      } else {
        client.write(msg);
        vscode.window.showInformationMessage(`Playing ${seqType}.${seqName}`);
      }
    }),

    vscode.commands.registerCommand('bleeder.stop', () => {
      if (client) {
        client.write('stop\n');
      }
    }),
  );
}

function deactivate() { }

module.exports = { activate, deactivate };
