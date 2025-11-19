local lines = {}
local job = vim.fn.jobstart({'/Users/jadercorrea/workspace/opensource/chuchu/test_raw_job.sh'}, {
  stdout_buffered = false,
  on_stdout = function(j, data)
    for _, line in ipairs(data or {}) do
      if line ~= '' then
        table.insert(lines, line)
        print('GOT:', line)
      end
    end
  end,
  on_exit = function()
    print('TOTAL:', #lines)
    vim.cmd('qall!')
  end,
})
print('JOB:', job)
vim.defer_fn(function() vim.cmd('qall!') end, 2000)
