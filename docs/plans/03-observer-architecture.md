# GPTCode Observer: Architecture & Vision

**Status:** 🌐 Marketing Site Complete | ❌ CLI Not Implemented  
**Last Updated:** 2025-12-01

## Executive Summary

**Current State:**
- ✅ **Marketing visualization** - Interactive demo on Jekyll homepage (100% complete)
- ❌ **CLI telemetry** - Not implemented (this doc describes future work)
- ❌ **Real-time WebSocket server** - Not implemented
- ❌ **Live dashboard** - Not implemented

**Vision:** Transform the GPTCode CLI from a black box into a **didactic observatory** through architectural decoupling and explainability.

## Core Principles

### 1. Observability as Education
- **Goal**: Not just show *what* CLI does, but *why* it makes decisions
- **Method**: Treat command flow as an observable system with attribution
- **Outcome**: Users understand AI decision-making process

### 2. Architectural Decoupling
```
┌─────────────────┐
│ CLI Core        │
│ (Instrumented)  │
└────────┬────────┘
         │ Events
         ▼
┌─────────────────┐       ┌──────────────────┐
│ State Tracers   │──────▶│ Visualization    │
│ (Data Layer)    │ JSON  │ Engine (View)    │
└─────────────────┘       └──────────────────┘
```

**Benefits**:
- Add new algorithms without rewriting visualizer
- Test tracers independently
- Multiple visualization front-ends possible

## Proposed: Tripartite Panel System

### Panel 1: Flow View (Current)
- Node graph with animation
- Force-directed or Sankey layout
- Real-time step highlighting

### Panel 2: Attribution View (New)
**Inspired by LIME/SHAP for XAI**

Shows **credit assignment** for each decision:

```
Decision: "Use PlannerAgent"
┌─────────────────────────────────┐
│ Contributing Factors:           │
├─────────────────────────────────┤
│ ████████████ 60% - Complexity   │
│ ██████ 30% - Multi-file change  │
│ ██ 10% - User preference        │
└─────────────────────────────────┘
```

**Implementation**:
- Capture decision inputs in trace data
- Calculate feature importance
- Visualize as stacked bar or force diagram

### Panel 3: Timeline View (New)
**Chronological execution with metrics**

```
t=0s    CommandParser ──────┐
                             │
t=0.5s  Intelligence ────────┤
                             ├─ Model: gpt-4o-mini
t=1.2s  PlannerAgent ────────┤   Cost: $0.0023
                             │
t=3.5s  EditorAgent ─────────┘
```

**Features**:
- Parallel vs sequential execution
- Cost/time breakdown
- Error recovery paths

## Data Architecture

### Trace Schema (Proposed)
```json
{
  "session_id": "uuid",
  "steps": [
    {
      "node": "CommandParser",
      "timestamp": 0,
      "duration_ms": 150,
      "cost": 0,
      "inputs": {
        "command": "gptcode do 'add login'",
        "context": {...}
      },
      "outputs": {
        "parsed_intent": "implement_feature",
        "complexity": 0.7
      },
      "decision": {
        "type": "route",
        "chosen": "IntelligenceSystem",
        "alternatives": ["DoCommand"],
        "attribution": {
          "complexity_score": 0.6,
          "multi_file_indicator": 0.3,
          "user_mode": 0.1
        }
      }
    }
  ]
}
```

### Instrumentation Points
Add to CLI codebase:

```go
// internal/observability/tracer.go
type Tracer interface {
    RecordDecision(node string, decision Decision)
    RecordMetrics(node string, metrics Metrics)
    RecordTransition(from, to string)
}

// Example usage in orchestrator
func (o *Orchestrator) Route(ctx Context) Agent {
    decision := o.intelligenceSystem.Decide(ctx)
    
    tracer.RecordDecision("IntelligenceSystem", Decision{
        Chosen: decision.Agent,
        Attribution: decision.Reasoning,
    })
    
    return decision.Agent
}
```

## Explainability Integration

### LIME-inspired Attribution
For each decision node:
1. **Identify features**: complexity, file count, code size, context
2. **Perturb inputs**: Vary features, observe output changes
3. **Calculate weights**: Which features most influenced decision?
4. **Visualize**: Show as contribution bars

### Implementation Phases

**Phase 1: Enhanced Tracing** (Week 1-2)
- [ ] Design trace schema
- [ ] Instrument key decision points
- [ ] Capture attribution data
- [ ] Export to JSON

**Phase 2: Tripartite Panels** (Week 3-4)
- [ ] Refactor observer.js for multiple views
- [ ] Implement Panel 2 (Attribution)
- [ ] Implement Panel 3 (Timeline)
- [ ] Add panel switching/layouts

**Phase 3: XAI Integration** (Week 5-6)
- [ ] Build feature extraction pipeline
- [ ] Implement attribution calculator
- [ ] Create interactive attribution viz
- [ ] Add "explain this decision" modal

**Phase 4: Interactive Tutorials** (Week 7-8)
- [ ] Create guided tour system
- [ ] Build example scenarios library
- [ ] Add step-by-step walkthroughs
- [ ] Community contribution templates

## Technical Stack

### Backend (Tracing)
- **Language**: Go (existing CLI)
- **Format**: JSON lines (for streaming)
- **Storage**: Local files + optional remote

### Frontend (Visualization)
- **Core**: D3.js v7 (force graphs, timelines)
- **UI**: Vanilla JS + CSS (keep lightweight)
- **Layout**: CSS Grid for tripartite panels
- **Interactions**: GSAP for smooth animations

### Explainability
- **Method**: Simplified LIME (local linear approximation)
- **Implementation**: Custom JS (no heavy ML libs needed)
- **Focus**: Decision trees, not black-box models

## Example: Explaining "Why PlannerAgent?"

```javascript
{
  "node": "IntelligenceSystem",
  "decision": "Use PlannerAgent",
  "attribution": {
    "features": [
      {
        "name": "Complexity Score",
        "value": 0.78,
        "weight": 0.60,
        "direction": "positive"
      },
      {
        "name": "Multi-file Change",
        "value": true,
        "weight": 0.30,
        "direction": "positive"
      },
      {
        "name": "Has Tests",
        "value": false,
        "weight": 0.10,
        "direction": "negative"
      }
    ]
  }
}
```

**Visualization**:
```
PlannerAgent chosen because:
━━━━━━━━━━━━ 60% High complexity (0.78)
━━━━━━ 30% Multiple files affected
━━ 10% No existing tests (-)
```

## Community Engagement

### Interactive Tutorials
- **Beginner**: "Your First Orchestrated Command"
- **Intermediate**: "Understanding AI Routing Decisions"
- **Advanced**: "Custom Agent Development"

### Open Visualization Data
- Publish anonymized trace datasets
- Enable community-built visualizations
- Visualization contest/showcase

## Success Metrics

1. **Educational Impact**
   - Users can explain why CLI chose a path
   - Reduced confusion about AI decisions
   
2. **Technical Extensibility**
   - Adding new agent requires <10 lines of tracer code
   - Visualization updates automatically
   
3. **Community Adoption**
   - 3+ community-contributed tutorials
   - External visualizations using trace data

## Next Steps

1. **This Week**: Implement enhanced trace schema
2. **Next Sprint**: Build Panel 2 (Attribution View)
3. **Month**: Complete Tripartite System
4. **Quarter**: Launch interactive tutorials

## References

- [Transformer Explainer](https://poloclub.github.io/transformer-explainer/)
- [LIME Paper](https://arxiv.org/abs/1602.04938)
- [Observable Plot](https://observablehq.com/plot/) for inspiration
- [D3 Force Directed](https://d3js.org/d3-force)

---

*This architecture ensures GPTCode Observatory evolves from a flow visualizer to a comprehensive educational platform for understanding AI-assisted development.*
