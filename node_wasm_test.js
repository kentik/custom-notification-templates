const fs = require('fs');
require('./dist/wasm_exec.js');

const template = fs.readFileSync('./templates/json-clean.json.tmpl', 'utf8');
const data = fs.readFileSync('./dist/evm.json', 'utf8');

global.fs = require('fs');

async function runWasm() {
  const go = new Go();

  const wasmBuffer = fs.readFileSync('./dist/renderer.wasm');
  const { instance } = await WebAssembly.instantiate(wasmBuffer, go.importObject);
  go.run(instance);

  try {
    if (typeof global.goTemplateRender !== 'function') {
      throw new Error('Go function not found on global scope!');
    }

    console.log('Calling Go function...');

    const result = global.goTemplateRender(template, data);

    console.log('Result from Go:', result);
  } catch (err) {
    console.error('Execution failed:', err);
  }

  process.exit(0);
}

runWasm();
