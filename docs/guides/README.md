# GPTCode Documentation Structure

## Organization

### Blog Posts (`_posts/`)
Time-stamped articles with opinions, announcements, and tutorials in context.

**Characteristics:**
- Dated content (YYYY-MM-DD prefix)
- Reflects state at time of writing
- May include opinions, comparisons, and motivations
- Examples: "Why GPTCode", "Groq Optimal Configs"

**Purpose:** Tell the story, provide context, share experiences

### Guides (`guides/`)
*Coming soon*

Reference documentation that stays current with the codebase.

**Characteristics:**
- Evergreen content (no dates)
- Updated with feature changes
- Technical reference and how-tos
- Examples: API reference, CLI commands, configuration schema

**Purpose:** Technical documentation that stays accurate

## Current State

Currently, all documentation lives in `_posts/`. This works but mixes temporal blog content with timeless technical docs.

Future guides will be extracted to `guides/` as the project matures.

### Available guides
- [Getting Started](./getting-started.md) - install, setup, quick start
- [Universal Feedback Capture](./feedback.md) - two‑keystroke feedback from any CLI

## Contributing

When documenting a feature:
1. **New feature announcement**: Add blog post in `_posts/`
2. **Feature documentation**: Update existing post or add to README
3. **Long-term reference**: (Future) Create guide in `guides/`

For now, enhance existing blog posts with technical details rather than creating new dated posts for every feature update.
