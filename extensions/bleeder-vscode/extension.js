// @ts-check
const vscode = require('vscode');
const net = require('net');

/** @param {vscode.ExtensionContext} context */
function activate(context) {
  let client = null;

  // CodeLens
  vscode.languages.registerCodeLensProvider('toml', {
    provideCodeLenses(document) {
      const lenses = [];
      const text = document.getText();

      // Match [lane.xxx] or [riff.xxx]
      const regex = /^\[(lane|riff)\.(\w+)\]/gm;
      let match;

      while ((match = regex.exec(text))) {
        const line = document.positionAt(match.index).line;
        const range = new vscode.Range(line, 0, line, 0);
        const seqName = match[2];
        const seqType = match[1];

        lenses.push(
          new vscode.CodeLens(range, {
            title: `Play`,
            command: 'bleeder.play',
            arguments: [seqName, seqType]
          })
        );
      }

      return lenses;
    }
  });

  // InlayHints
  vscode.languages.registerInlayHintsProvider('toml', {
    provideInlayHints(document, range) {
      const hints = [];

      // Find @seq:args patterns
      const regex = /@(\w+):(\w+)/g;
      let match;

      while ((match = regex.exec(document.getText()))) {
        const pos = document.positionAt(match.index + match[0].length);

        hints.push({
          position: pos,
          label: ' [4 beats, tune: d1]',
          kind: vscode.InlayHintKind.Type,
          paddingLeft: true
        });
      }

      return hints;
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
