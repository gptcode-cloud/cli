vim.cmd([[set runtimepath+=~/.config/nvim/pack/plugins/start/chuchu]])

local M = require('chuchu')
M.setup({})

vim.schedule(function()
  M.toggle_chat()
  vim.wait(200)
  
  local buf = vim.api.nvim_get_current_buf()
  local lc = vim.api.nvim_buf_line_count(buf)
  vim.api.nvim_buf_set_lines(buf, lc-1, lc, false, {'👤 | Analyse codebase and add pix payment'})
  
  local original_chansend = vim.fn.chansend
  vim.fn.chansend = function(job, data)
    print('[STDIN LENGTH]', #data)
    print('[STDIN CONTENT]', data)
    local parsed = vim.fn.json_decode(data)
    print('[MESSAGES]', #parsed.messages)
    for i, msg in ipairs(parsed.messages) do
      print('  [' .. i .. '] ' .. msg.role .. ': ' .. msg.content:sub(1, 50))
    end
    return original_chansend(job, data)
  end
  
  M.send_message_from_buffer()
  
  vim.wait(5000)
  vim.cmd('qall!')
end)
