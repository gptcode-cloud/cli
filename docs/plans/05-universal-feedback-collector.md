# Universal Feedback Collector

**Priority**: 5 (Nice to have)  
**Status**: Partially implemented (GPTCode-only feedback exists)  
**Goal**: Capture feedback from ANY CLI (Gemini, Warp, Claude, etc.), not just GPTCode

## Current State

✅ **What exists:**
- `gptcode feedback submit` - Submit feedback to GPTCode
- `gptcode feedback hook install` - Shell hooks for Ctrl+g → mark suggestion
- Feedback collection works for GPTCode's own outputs
- ML training pipeline consumes feedback

❌ **What's missing:**
- Cannot capture feedback from other CLIs (Gemini AI, Warp AI, Claude Code, Cursor)
- Cannot capture corrections from interactive sessions with external tools
- No universal "wrong → correct" capture mechanism

## Problem

Users interact with MANY AI coding assistants:
- Gemini AI (Google)
- Warp AI (built into terminal)
- Claude Code (Anthropic)
- Cursor (VSCode fork)
- GitHub Copilot

**Current limitation**: Feedback only works for `chu` commands. If user gets bad suggestion from Warp AI or Gemini, that learning is lost.

**Vision**: Capture corrections from ANY AI tool → Feed into GPTCode's ML training

## Proposed Solution

### Phase 1: Universal Shell Hook

Extend current Ctrl+g hook to work with ANY command output:

```bash
# Current (only works after chu commands)
$ gptcode chat "install docker"
> sudo apt install docker  # Wrong for macOS
<Ctrl+g>  # Marks this
$ brew install docker  # Corrected
# ✅ Feedback captured

# Proposed (works with ANY command)
$ warp-ai "install docker"
> sudo apt install docker  # Wrong for macOS
<Ctrl+g>  # Marks this
$ brew install docker  # Corrected
# ✅ Feedback captured to GPTCode
```

**Implementation:**
- Detect ANY command that produces output
- Allow Ctrl+g to mark ANY terminal line (not just chu outputs)
- Track: `source_cli` (warp, gemini, claude, cursor, chu)
- Store feedback with source attribution

### Phase 2: Browser Extension (Optional)

Capture corrections from web UIs:

```javascript
// Gemini AI web interface
1. User gets bad suggestion in textarea
2. User edits it
3. Browser extension detects edit
4. Sends to gptcode feedback submit --source=gemini-web
```

**Targets:**
- Gemini AI (gemini.google.com)
- ChatGPT Code (chat.openai.com)
- Claude (claude.ai)

### Phase 3: Editor Integration (Optional)

Capture from Cursor, VSCode Copilot:

```
1. Copilot suggests code
2. User edits suggestion
3. Extension captures diff
4. Sends to gptcode feedback
```

## Technical Design

### Enhanced Feedback Schema

```json
{
  "source": "warp-ai",  // NEW: warp-ai, gemini, claude, chu, cursor
  "source_version": "0.2024.11.26.08.17.stable_02",
  "agent": "unknown",  // chu-specific, null for external
  "sentiment": "bad",
  "kind": "command",
  "wrong": "sudo apt install docker",
  "correct": "brew install docker",
  "context": {
    "os": "darwin",
    "shell": "zsh",
    "prompt": "install docker"
  }
}
```

### Shell Hook Enhancement

**File**: Hook scripts (zsh/bash/fish)

```bash
# Current: Only captures after chu commands
# New: Capture after ANY command

function universal_feedback_mark() {
  # Save current command line (any command, not just chu)
  local current_line=$(commandline -b)  # fish
  # or: local current_line=$BUFFER      # zsh
  
  # Detect source CLI from command history
  local source_cli=$(detect_source_from_history)  # warp-ai, gemini, etc.
  
  echo "$current_line" > ~/.gptcode/feedback/pending_wrong
  echo "$source_cli" > ~/.gptcode/feedback/pending_source
}

bind \cg universal_feedback_mark  # Ctrl+g
```

**Detection heuristics:**
- Check command history for patterns (warp-ai, gemini-cli, etc.)
- Check environment variables (WARP_SESSION, GEMINI_SESSION)
- Fallback: `source=unknown`

### Privacy & Consent

**Important**: User must opt-in to universal collection

```bash
# Enable universal feedback (default: off)
gptcode feedback config --universal=true

# Show what's being collected
gptcode feedback config --show

# Disable
gptcode feedback config --universal=false
```

**Data handling:**
- Only command text + corrections (no secrets)
- Local storage first (`~/.gptcode/feedback/`)
- User reviews before submission (optional)
- Clear privacy policy

## Implementation Phases

### Phase 1: Universal Shell Hook (Week 1-2)
- [ ] Enhance shell hooks to capture ANY command
- [ ] Add source CLI detection
- [ ] Update feedback schema with `source` field
- [ ] Add privacy opt-in flag
- [ ] Test with Warp AI, Gemini CLI

### Phase 2: ML Training Integration (Week 2)
- [ ] Update ML pipeline to handle multi-source feedback
- [ ] Weight feedback by source (chu > external)
- [ ] Add source attribution to training data
- [ ] Retrain intent classifier with external feedback

### Phase 3: Browser Extension (Week 3-4, Optional)
- [ ] Chrome extension for Gemini AI
- [ ] Capture textarea edits
- [ ] Send to gptcode feedback API
- [ ] Firefox support

### Phase 4: Editor Integration (Week 5+, Optional)
- [ ] VSCode extension
- [ ] Cursor integration
- [ ] Copilot diff capture

## Success Criteria

- [ ] Ctrl+g works after Warp AI suggestions
- [ ] Feedback captures `source=warp-ai`
- [ ] ML training consumes multi-source feedback
- [ ] Privacy opt-in working
- [ ] No secrets leaked in feedback

## Benefits

1. **More training data**: Learn from ALL AI interactions, not just GPTCode
2. **Competitive intelligence**: See where others fail
3. **Faster improvement**: More corrections = better models
4. **Unique feature**: No other AI tool does this

## Risks

- **Privacy concerns**: Must be transparent about data collection
- **Complexity**: Detecting external CLIs is heuristic-based
- **Browser security**: Extensions need careful review
- **Marginal value**: Most corrections might be GPTCode-specific anyway

## Priority Rationale

**Why Priority 5 (Nice to have)?**
- Current feedback system works well for GPTCode
- Most users primarily use GPTCode (not other tools)
- Implementation complexity high (shell detection, browser extensions)
- Privacy/security considerations significant
- Better to focus on improving GPTCode quality first

**When to prioritize:**
- After user-friendly commands (P1)
- After programmatic tools (P2)
- After core features are polished
- When users explicitly request cross-tool learning
