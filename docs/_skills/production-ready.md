---
layout: skill
title: Production Ready
name: production-ready
category: product
description: Ship robust features with proper error handling, feature flags, health checks, and observability built-in.
---

# Production Ready

Guidelines for shipping features that are resilient, observable, and safely deployable.

## When to Activate

- Preparing features for production
- Implementing error handling
- Adding observability
- Setting up feature flags

## Error Boundaries

### React error boundaries

```tsx
import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log to monitoring service
    console.error('Error caught:', error, errorInfo);
    this.props.onError?.(error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div role="alert">
          <h2>Something went wrong</h2>
          <button onClick={() => this.setState({ hasError: false })}>
            Try again
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

// Usage - wrap critical sections
<ErrorBoundary 
  fallback={<FallbackUI />}
  onError={(e) => Sentry.captureException(e)}
>
  <CriticalFeature />
</ErrorBoundary>
```

### Async error handling

```typescript
// GOOD - graceful degradation
async function fetchUserData(userId: string) {
  try {
    const response = await api.get(`/users/${userId}`);
    return { data: response.data, error: null };
  } catch (error) {
    // Log but don't crash
    console.error('Failed to fetch user:', error);
    Sentry.captureException(error);
    
    // Return fallback
    return { 
      data: null, 
      error: error instanceof Error ? error.message : 'Unknown error' 
    };
  }
}

// Usage
const { data, error } = await fetchUserData(id);
if (error) {
  showNotification('Using cached data', 'warning');
  return getCachedUser(id);
}
```

## Feature Flags

### Simple feature flag system

```typescript
// flags.ts
type FeatureFlag = {
  enabled: boolean;
  percentage?: number; // Gradual rollout
  allowList?: string[]; // Specific users
};

const FLAGS: Record<string, FeatureFlag> = {
  NEW_CHECKOUT: { enabled: false, percentage: 10 },
  DARK_MODE: { enabled: true },
  BETA_FEATURES: { enabled: true, allowList: ['user_123', 'user_456'] },
};

export function isFeatureEnabled(
  flag: string, 
  userId?: string
): boolean {
  const config = FLAGS[flag];
  if (!config) return false;
  if (!config.enabled) return false;
  
  // Check allowlist
  if (config.allowList && userId) {
    return config.allowList.includes(userId);
  }
  
  // Percentage rollout (deterministic by userId)
  if (config.percentage !== undefined && userId) {
    const hash = hashCode(userId + flag);
    return (hash % 100) < config.percentage;
  }
  
  return config.enabled;
}

// Usage
if (isFeatureEnabled('NEW_CHECKOUT', user.id)) {
  return <NewCheckout />;
}
return <LegacyCheckout />;
```

### Feature flag with remote config

```typescript
// Using environment-based flags
const config = {
  features: {
    newDashboard: process.env.FEATURE_NEW_DASHBOARD === 'true',
    aiAssistant: process.env.FEATURE_AI_ASSISTANT === 'true',
  },
};

// Or fetch from remote
async function loadFeatureFlags() {
  const response = await fetch('/api/feature-flags');
  return response.json();
}
```

## Health Checks

### HTTP health endpoints

```typescript
// Express.js
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/health/ready', async (req, res) => {
  const checks = await Promise.allSettled([
    checkDatabase(),
    checkRedis(),
    checkExternalAPI(),
  ]);
  
  const results = {
    database: checks[0].status === 'fulfilled',
    redis: checks[1].status === 'fulfilled',
    externalAPI: checks[2].status === 'fulfilled',
  };
  
  const allHealthy = Object.values(results).every(Boolean);
  
  res.status(allHealthy ? 200 : 503).json({
    status: allHealthy ? 'ready' : 'degraded',
    checks: results,
    timestamp: new Date().toISOString(),
  });
});

app.get('/health/live', (req, res) => {
  // Simple liveness check - is the process running?
  res.json({ status: 'alive' });
});
```

### Kubernetes probes

```yaml
# deployment.yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 3000
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Graceful Degradation

### Circuit breaker pattern

```typescript
class CircuitBreaker {
  private failures = 0;
  private lastFailure?: Date;
  private state: 'closed' | 'open' | 'half-open' = 'closed';
  
  constructor(
    private threshold = 5,
    private resetTimeout = 30000
  ) {}

  async execute<T>(fn: () => Promise<T>, fallback: T): Promise<T> {
    if (this.state === 'open') {
      if (Date.now() - this.lastFailure!.getTime() > this.resetTimeout) {
        this.state = 'half-open';
      } else {
        return fallback;
      }
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      return fallback;
    }
  }

  private onSuccess() {
    this.failures = 0;
    this.state = 'closed';
  }

  private onFailure() {
    this.failures++;
    this.lastFailure = new Date();
    if (this.failures >= this.threshold) {
      this.state = 'open';
    }
  }
}

// Usage
const paymentCircuit = new CircuitBreaker();
const result = await paymentCircuit.execute(
  () => paymentAPI.process(order),
  { success: false, message: 'Payment temporarily unavailable' }
);
```

## Structured Logging

### Use consistent log format

```typescript
import pino from 'pino';

const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  formatters: {
    level: (label) => ({ level: label }),
  },
});

// GOOD - structured, searchable
logger.info({
  event: 'order_created',
  orderId: order.id,
  userId: user.id,
  total: order.total,
  items: order.items.length,
}, 'Order created successfully');

// BAD - unstructured
console.log(`Order ${order.id} created for user ${user.id}`);
```

### Include request context

```typescript
// Middleware to add request ID
app.use((req, res, next) => {
  req.requestId = req.headers['x-request-id'] || crypto.randomUUID();
  req.log = logger.child({ requestId: req.requestId });
  next();
});

// Use in handlers
app.post('/orders', async (req, res) => {
  req.log.info({ body: req.body }, 'Processing order');
  // ...
  req.log.info({ orderId: order.id }, 'Order created');
});
```

## Timeouts & Retries

### Always set timeouts

```typescript
// GOOD - explicit timeout
const response = await fetch(url, {
  signal: AbortSignal.timeout(5000), // 5s timeout
});

// With retry logic
async function fetchWithRetry(
  url: string,
  options: RequestInit = {},
  retries = 3
): Promise<Response> {
  for (let i = 0; i < retries; i++) {
    try {
      const response = await fetch(url, {
        ...options,
        signal: AbortSignal.timeout(5000),
      });
      if (response.ok) return response;
    } catch (error) {
      if (i === retries - 1) throw error;
      // Exponential backoff
      await new Promise(r => setTimeout(r, Math.pow(2, i) * 1000));
    }
  }
  throw new Error('Max retries exceeded');
}
```

## Deployment Safety

### Pre-deployment checklist

```markdown
## Before Deploying

- [ ] All tests passing
- [ ] Feature flag configured (if new feature)
- [ ] Rollback plan documented
- [ ] Monitoring alerts configured
- [ ] Database migrations tested
- [ ] API backwards compatible
- [ ] Load tested for expected traffic
- [ ] Security review completed
```
