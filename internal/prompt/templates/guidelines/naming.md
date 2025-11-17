# Naming Conventions

Naming is one of the most critical aspects of code quality. Names must reveal intent, be searchable, and avoid encoding.

## Core Principles (Clean Code, Code Complete)

**Be Descriptive and Intent-Revealing**
A name should answer:
- Why does it exist?
- What does it do?
- How is it used?

```python
d = 42
days_since_last_login = 42
```

**Choose Searchable Names**
- Use full, pronounceable words
- Avoid single-letter names (except loop indices in small scopes)
- Avoid cryptic abbreviations

```typescript
const yyyymmdstr = format(new Date(), 'yyyy/MM/dd')
const currentDateFormatted = format(new Date(), 'yyyy/MM/dd')
```

**Avoid Encoding Type (Anti-Hungarian Notation)**
- Don't prefix with data types
- Modern IDEs/editors show types on hover

```go
strName := "John"
iAge := 30

name := "John"
age := 30
```

## Naming by Role

**Modules/Classes/Types** → Nouns representing domain concepts:
- `InvoiceCalculator`, `OrderProcessor`, `UserRepository`
- NOT: `InvoiceHelper`, `OrderManager`, `UserUtils`

**Functions/Methods** → Verbs representing operations:
- `calculate_total`, `validate_order`, `build_response`
- `is_valid`, `has_permission`, `can_process` (predicates)
- NOT: `do_stuff`, `handle`, `process` (too vague)

**Variables** → Nouns, should reveal what they hold:
- `active_users`, `total_amount`, `error_message`
- NOT: `data`, `info`, `temp`, `x`

**Constants** → ALL_CAPS or camelCase depending on language:
- `MAX_RETRY_ATTEMPTS`, `DEFAULT_TIMEOUT_MS`
- Should read like prose: `if retry_count > MAX_RETRY_ATTEMPTS`

## Context and Scope

**Use longer names for wider scopes:**
```ruby
def calc(items)
end

def calculate_invoice_total(line_items)
end
```

**Shorter names OK in tight scopes:**
```elixir
Enum.map(items, fn i -> i.price * i.quantity end)

Enum.map(line_items, fn item -> 
  item.price * item.quantity 
end)
```

## Red Flags

Challenge any name that includes:
- `manager`, `helper`, `utils`, `handler` → Usually too vague
- `data`, `info`, `object` → Says nothing about content
- `do_`, `perform_`, `execute_` → Redundant verb prefix
- `temp`, `tmp`, `x`, `foo` → Placeholder that stuck around

## Validation Process

**Before accepting any name, ask:**
1. Can I pronounce it in a conversation?
2. Can I search for it across the codebase?
3. Does it reveal intent without needing a comment?
4. Would a new team member understand it?

If the answer to any is "no", propose a better name.

## Comments

**NEVER write useless comments.**

- Comments that simply restate what the code does are noise: `// Memory file`, `// Loop through items`, `// Set variable`.
- If the code is clear, no comment is needed.
- Only add comments for:
  - Non-obvious **why** decisions or trade-offs
  - Complex algorithms that benefit from explanation
  - TODOs or known limitations
- When in doubt, improve the code clarity instead of adding a comment.
