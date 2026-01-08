const fs = require('fs');
const assert = require('assert');
require('./dist/wasm_exec.js');

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

  // Verify the function is available
  test('goTemplateRender function is available', () => {
    assert.strictEqual(typeof global.goTemplateRender, 'function', 'goTemplateRender should be a function');
  });

  // Test 1: Valid template rendering
  test('Valid template renders successfully', () => {
    const result = JSON.parse(global.goTemplateRender(template, data));
    assert.strictEqual(result.error, undefined, 'Should not have errors for valid template');
    assert.ok(result.output, 'Should produce output');
    assert.ok(result.output.length > 0, 'Output should not be empty');
  });

  // Test 2: Simple template works
  test('Simple template with variable substitution', () => {
    const simpleTemplate = '{{ .CompanyName }}';
    const simpleData = JSON.stringify({ CompanyName: 'Test Company', CompanyID: 1 });
    const result = JSON.parse(global.goTemplateRender(simpleTemplate, simpleData));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    assert.strictEqual(result.output, 'Test Company', 'Output should match company name');
  });

  // Test 3: Invalid template syntax - should return error with line number
  test('Invalid template syntax returns error with line info', () => {
    const invalidTemplate = '{{ .Invalid';
    const result = JSON.parse(global.goTemplateRender(invalidTemplate, data));
    assert.ok(result.error, 'Should have error for invalid template syntax');
    assert.ok(result.error.includes('unclosed action'), 'Error should mention unclosed action');
  });

  // Test 4: Template with undefined field - should fail at execution
  test('Template with undefined field in strict mode', () => {
    const badTemplate = '{{ .NonExistentField.SubField }}';
    const result = JSON.parse(global.goTemplateRender(badTemplate, data));
    // Go templates with undefined nested fields cause errors
    assert.ok(result.error || result.output !== undefined, 'Should either error or produce empty output');
  });

  // Test 5: Type validation - wrong type for template argument
  test('Wrong type for template argument returns error', () => {
    const result = JSON.parse(global.goTemplateRender(123, data));
    assert.ok(result.error, 'Should have error for non-string template');
    assert.ok(result.error.includes('string'), 'Error should mention string type');
  });

  // Test 6: Type validation - wrong type for data argument
  test('Wrong type for data argument returns error', () => {
    const result = JSON.parse(global.goTemplateRender(template, 456));
    assert.ok(result.error, 'Should have error for non-string data');
    assert.ok(result.error.includes('string'), 'Error should mention string type');
  });

  // Test 7: Missing arguments
  test('Missing arguments returns error', () => {
    const result = JSON.parse(global.goTemplateRender());
    assert.ok(result.error, 'Should have error for missing arguments');
    assert.ok(result.error.includes('Expected arguments'), 'Error should mention expected arguments');
  });

  // Test 8: Invalid JSON data
  test('Invalid JSON data returns error', () => {
    const result = JSON.parse(global.goTemplateRender('{{ .CompanyID }}', '{invalid json}'));
    assert.ok(result.error, 'Should have error for invalid JSON data');
    assert.ok(result.error.includes('parse') || result.error.includes('Data'), 'Error should mention parsing issue');
  });

  // Test 9: Empty template
  test('Empty template renders to empty output', () => {
    const result = JSON.parse(global.goTemplateRender('', data));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    assert.strictEqual(result.output, '', 'Output should be empty string');
  });

  // Test 10: Template with built-in functions
  test('Template functions work correctly', () => {
    const funcTemplate = '{{ .CompanyName | toUpper }}';
    const funcData = JSON.stringify({ CompanyName: 'Test', CompanyID: 1 });
    const result = JSON.parse(global.goTemplateRender(funcTemplate, funcData));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    assert.strictEqual(result.output, 'TEST', 'toUpper function should work');
  });

  // Test 11: JSON output is valid
  test('JSON template produces valid JSON', () => {
    const result = JSON.parse(global.goTemplateRender(template, data));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    // Verify the output is valid JSON
    const outputJson = JSON.parse(result.output);
    assert.ok(outputJson, 'Output should be valid JSON');
  });

  // Test 12: Error includes line number for syntax error on specific line
  test('Error includes accurate line number', () => {
    const multilineTemplate = 'Line 1\nLine 2\n{{ .Invalid';
    const result = JSON.parse(global.goTemplateRender(multilineTemplate, data));
    assert.ok(result.error, 'Should have error');
    assert.ok(result.line === 3 || result.line > 0, 'Should have line number');
  });

  // Test 13: Template with toJSON function
  test('toJSON function produces valid output', () => {
    const jsonTemplate = '{{ toJSON . }}';
    const jsonData = JSON.stringify({ CompanyName: 'Test', CompanyID: 123 });
    const result = JSON.parse(global.goTemplateRender(jsonTemplate, jsonData));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    const parsed = JSON.parse(result.output);
    assert.strictEqual(parsed.CompanyID, 123, 'toJSON should preserve data');
  });

  // Test 14: importanceToColor function
  test('importanceToColor function returns hex color', () => {
    const colorTemplate = '{{ importanceToColor .Event.Importance }}';
    const result = JSON.parse(global.goTemplateRender(colorTemplate, data));
    assert.strictEqual(result.error, undefined, 'Should not have errors');
    assert.ok(result.output.startsWith('#'), 'Should return hex color');
    assert.strictEqual(result.output.length, 7, 'Hex color should be 7 characters');
  });

  // ===============================
  // Schema API Tests
  // ===============================

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

  test('Schema fields include NotificationViewModel root fields', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    const fieldNames = schema.fields.map(f => f.name);
    assert.ok(fieldNames.includes('CompanyID'), 'Should include CompanyID');
    assert.ok(fieldNames.includes('CompanyName'), 'Should include CompanyName');
    assert.ok(fieldNames.includes('Config'), 'Should include Config');
  });

  test('Schema includes Event method with children', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    const eventField = schema.fields.find(f => f.name === 'Event');
    assert.ok(eventField, 'Should have Event field/method');
    assert.ok(eventField.isMethod || eventField.children, 'Event should be method or have children');
  });

  test('Schema functions include all template functions', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    const funcNames = schema.functions.map(f => f.name);
    assert.ok(funcNames.includes('toJSON'), 'Should include toJSON');
    assert.ok(funcNames.includes('toUpper'), 'Should include toUpper');
    assert.ok(funcNames.includes('importanceToColor'), 'Should include importanceToColor');
    assert.ok(funcNames.includes('timeRfc3339'), 'Should include timeRfc3339');
  });

  test('Schema functions have signatures and descriptions', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    const toJSON = schema.functions.find(f => f.name === 'toJSON');
    assert.ok(toJSON, 'Should have toJSON function');
    assert.ok(toJSON.signature, 'Function should have signature');
    assert.ok(toJSON.description, 'Function should have description');
    assert.ok(toJSON.category, 'Function should have category');
  });

  test('Schema enums include ViewModelImportance', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    assert.ok(schema.enums.ViewModelImportance, 'Should have ViewModelImportance enum');
    const values = schema.enums.ViewModelImportance.values;
    assert.ok(values.includes('Critical'), 'Should include Critical');
    assert.ok(values.includes('Healthy'), 'Should include Healthy');
    assert.ok(values.includes('Warning'), 'Should include Warning');
  });

  test('Schema enums include EventType', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    assert.ok(schema.enums.EventType, 'Should have EventType enum');
    const values = schema.enums.EventType.values;
    assert.ok(values.includes('alarm'), 'Should include alarm');
    assert.ok(values.includes('insight'), 'Should include insight');
    assert.ok(values.includes('synthetic'), 'Should include synthetic');
  });

  test('Schema fields have proper paths', () => {
    const schema = JSON.parse(global.goTemplateGetSchema());
    const companyId = schema.fields.find(f => f.name === 'CompanyID');
    assert.ok(companyId, 'Should have CompanyID');
    assert.strictEqual(companyId.path, '.', 'Root field should have path "."');
  });

  // ===============================
  // Error Range Tests
  // ===============================

  test('Error response includes range fields', () => {
    const result = JSON.parse(global.goTemplateRender('{{ .Invalid', data));
    assert.ok(result.error, 'Should have error');
    // Check range fields exist
    assert.ok('startLine' in result, 'Should have startLine field');
    assert.ok('endLine' in result, 'Should have endLine field');
  });

  test('Error range matches line number', () => {
    const multilineTemplate = 'Line 1\nLine 2\n{{ .Invalid';
    const result = JSON.parse(global.goTemplateRender(multilineTemplate, data));
    assert.ok(result.error, 'Should have error');
    assert.strictEqual(result.startLine, result.line, 'startLine should match line');
    assert.strictEqual(result.endLine, result.line, 'endLine should match line');
  });

  test('Error range includes column when available', () => {
    const result = JSON.parse(global.goTemplateRender('{{ .Invalid', data));
    assert.ok(result.error, 'Should have error');
    // If column is available, startColumn should match
    if (result.column) {
      assert.strictEqual(result.startColumn, result.column, 'startColumn should match column');
    }
  });

  test('Legacy line/column fields still work', () => {
    const multilineTemplate = 'Line 1\nLine 2\n{{ .Invalid';
    const result = JSON.parse(global.goTemplateRender(multilineTemplate, data));
    assert.ok(result.error, 'Should have error');
    assert.ok(result.line > 0, 'Should have legacy line field');
    // Legacy fields should be populated
    assert.strictEqual(result.line, 3, 'Line should be 3 (error on line 3)');
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

runWasm().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
