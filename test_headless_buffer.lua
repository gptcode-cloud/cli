vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])
local M = require('chuchu')
M.setup({})

local got_stdout = false
local parsed_events = 0
local lines_seen = {}

local orig_jobstart = vim.fn.jobstart
vim.fn.jobstart = function(cmd, opts)
  local wrapped = vim.deepcopy(opts)
  local old_out = opts.on_stdout
  wrapped.on_stdout = function(j, data, n)
    if data then
      for _, line in ipairs(data) do
        if line ~= '' then
          got_stdout = true
          table.insert(lines_seen, line)
          local m = line:match("__EVENT__(.+)__EVENT__")
          if m then parsed_events = parsed_events + 1 end
        end
      end
    end
    if old_out then old_out(j, data, n) end
  end
  return orig_jobstart(cmd, wrapped)
end

vim.schedule(function()
  M.toggle_chat()
  vim.wait(150)
  local buf = vim.api.nvim_get_current_buf()
  local lc = vim.api.nvim_buf_line_count(buf)
  local last = vim.api.nvim_buf_get_lines(buf, lc-1, lc, false)[1] or ''
  if not last:match('^👤 %|') then
    vim.api.nvim_buf_set_lines(buf, lc, lc, false, {'👤 | '})
    lc = vim.api.nvim_buf_line_count(buf)
  end
  vim.api.nvim_buf_set_lines(buf, lc-1, lc, false, {'👤 | Analyse codebase and add pix payment'})
  M.send_message_from_buffer()
  vim.wait(8000)
  print('STDOUT:', got_stdout)
  print('LINES:', #lines_seen)
  print('EVENTS:', parsed_events)
  if not got_stdout then print('FAIL_STDOUT') end
  if parsed_events == 0 then
    local first = lines_seen[1] or ''
    io.stdout:write('FIRST:' .. first .. '\n')
    print('FAIL_PARSE')
  end
  vim.cmd('qall!')
end)
