const fs = require('fs');
const assert = require('assert');
require('../dist/wasm_exec.js');

const template = fs.readFileSync('./templates/json-clean.json.tmpl', 'utf8');
const data = fs.readFileSync('./pkg/render/fixtures/alarm.json', 'utf8');

global.fs = require('fs');

let testsRun = 0;
let testsPassed = 0;

function test(name, fn) {
  testsRun++;
  try {
    fn();
    testsPassed++;
    console.log(`  PASS: ${name}`);
  } catch (err) {
    console.error(`  FAIL: ${name}`);
    console.error(`        ${err.message}`);
    throw err;
  }
}
async function runWasm() {
  const go = new Go();

  const wasmBuffer = fs.readFileSync('./dist/renderer.wasm');
  const { instance } = await WebAssembly.instantiate(wasmBuffer, go.importObject);
  go.run(instance);

  console.log('WASM Integration Tests');
  console.log('======================\n');

  test('goTemplateRender function is available', () => {
    assert.strictEqual(typeof global.goTemplateRender, 'function', 'goTemplateRender should be a function');
  });

  test('Simple template with variable substitution', () => {
    const simpleTemplate = '{{ .CompanyName }}';
    const simpleData = JSON.stringify({ CompanyName: 'Test Company', CompanyID: 1 });
    const result = JSON.parse(global.goTemplateRender(simpleTemplate, simpleData));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    assert.strictEqual(result.output, 'Test Company', 'Output should match company name');
  });

  test('Error includes accurate line number', () => {
    const multilineTemplate = 'Line 1\nLine 2\n{{ .Invalid';
    const result = JSON.parse(global.goTemplateRender(multilineTemplate, data));
    assert.ok(result.error, 'Should have error');
    assert.ok(result.line === 3 || result.line > 0, 'Should have line number');
  });

  test('goTemplateGetSchema function is available', () => {
    assert.strictEqual(typeof global.goTemplateGetSchema, 'function', 'goTemplateGetSchema should be a function');
  });

  test('Schema returns valid JSON with required sections', () => {
    const schemaStr = global.goTemplateGetSchema();
    const schema = JSON.parse(schemaStr);
    assert.ok(schema.fields, 'Schema should have fields array');
    assert.ok(schema.functions, 'Schema should have functions array');
    assert.ok(schema.enums, 'Schema should have enums object');
  });

  console.log('\n======================');
  console.log(`Tests: ${testsPassed}/${testsRun} passed`);

  if (testsPassed !== testsRun) {
    console.error('Some tests failed!');
    process.exit(1);
  }

  console.log('All tests passed!');
  process.exit(0);
}

runWasm().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});
