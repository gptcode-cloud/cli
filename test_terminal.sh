#!/bin/bash
rm -f /tmp/nvim_test_result.txt

nvim -c "luafile ~/workspace/opensource/gptcode/test_terminal.lua" /tmp/nvim_test_result.txt

cat /tmp/nvim_test_result.txt
