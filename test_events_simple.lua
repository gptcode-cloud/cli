vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local M = require('chuchu')
M.setup({})

print("[TEST] Testing event handler")

local test_events = {
  {json = '{"type":"message","data":{"content":"Test message"}}', expected = "message"},
  {json = '{"type":"status","data":{"status":"Working..."}}', expected = "status"},
  {json = '{"type":"confirm","data":{"prompt":"Continue?","id":"test"}}', expected = "confirm"},
}

local function mock_vim_ui_input(opts, callback)
  print("[TEST] vim.ui.input called:", opts.prompt)
  callback("y")
end

vim.ui.input = mock_vim_ui_input

for i, test in ipairs(test_events) do
  print(string.format("\n[TEST] Test %d: %s", i, test.expected))
  local ok, err = pcall(function()
    M.handle_tool_event(test.json, 1)
  end)
  if ok then
    print(string.format("[TEST] ✓ Event %s handled successfully", test.expected))
  else
    print(string.format("[TEST] ✗ Event %s failed: %s", test.expected, err))
  end
end

print("\n[TEST] Complete")
vim.cmd('qall!')
