# magicCommon Framework Lifecycle Improvement Plan

Status: implemented

Last Updated: 2026-06-07

## Purpose

This document defines a framework-level improvement plan for `magicCommon/framework`.
It focuses on generic lifecycle, service, and plugin semantics that are useful for
applications built on top of magicCommon.

The goal is to make `Application`, `Service`, `initiator`, and `module` boundaries
clearer, safer, and easier to compose without moving application-specific business
logic into the framework.

## Implementation Status

Implemented on 2026-06-07.

Code changes:

- `framework/plugin/common/util.go`
- `framework/plugin/initiator/initiator.go`
- `framework/plugin/module/module.go`
- `framework/application/application.go`
- `framework/service/lifecycle.go`

Test coverage:

- plugin registration validation, duplicate IDs, invalid method signatures
- explicit plugin interface invocation and legacy reflection compatibility
- manager-level setup rollback order
- Application run-before-startup, repeated startup, shutdown-before-startup,
  failed startup cleanup, options startup, injected runtime ownership
- Lifecycle adapter through real Application startup / run / shutdown

Validation completed:

```bash
GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor go test ./... -count=1
```

## Current Design

### Application

Current code: `framework/application/application.go`

`Application` owns the process-wide runtime container:

- default `event.Hub`
- default `task.BackgroundRoutine`
- default configuration manager initialization
- injected `service.Service`
- shutdown and runtime reconstruction

Current interface:

```go
type Application interface {
    Startup(ctx context.Context, service service.Service) *cd.Error
    Run(ctx context.Context) *cd.Error
    Shutdown(ctx context.Context)
    EventHub() event.Hub
    BackgroundRoutine() task.BackgroundRoutine
}
```

### Service

Current code: `framework/service/service.go`

`Service` is the lifecycle coordinator between Application and plugins.

Current interface:

```go
type Service interface {
    Startup(ctx context.Context, serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error
    Run(ctx context.Context) *cd.Error
    Shutdown(ctx context.Context)
}
```

Current `DefaultService` lifecycle order:

1. `initiator.Setup`
2. configured dependency checks
3. `module.Setup`
4. `initiator.Run`
5. `module.Run`
6. `module.Teardown`
7. `initiator.Teardown`

### Plugin

Current code:

- `framework/plugin/common/util.go`
- `framework/plugin/initiator/initiator.go`
- `framework/plugin/module/module.go`

Current plugin manager behavior:

- plugin must be a pointer
- plugin must implement `ID()` and `Run(ctx)`
- `Weight`, `Setup`, and `Teardown` are optional
- plugins are ordered by ascending `Weight`
- `Teardown` runs in reverse order
- `initiator` supports typed `GetEntity`
- `module` exposes registration and lifecycle only

## Problems To Solve

### 1. Registration errors can be hidden

`initiator.Register` and `module.Register` currently ignore plugin manager
registration errors. Duplicate IDs, non-pointer plugins, nil plugins, or missing
required methods can fail silently.

This makes invalid runtime assembly hard to detect during application startup.

### 2. Setup rollback is owned by callers, not by PluginMgr

`DefaultService` calls `module.Teardown` or `initiator.Teardown` when higher-level
setup fails, but `PluginMgr.Setup` does not track which plugins completed setup.

If plugin N fails after plugins 1..N-1 completed setup, teardown currently depends
on the caller tearing down the entire manager. A manager-level rollback would be
more precise and safer for future reuse.

### 3. Reflection-only plugin contract hides compile-time intent

The plugin manager currently uses reflection to discover `ID`, `Weight`, `Setup`,
`Run`, and `Teardown`.

Reflection preserves loose compatibility, but new code has no compile-time
contract for plugin shape, setup signature, teardown signature, or lifecycle
intent.

### 4. Application startup options are global and implicit

`Application.Startup` initializes configuration with an empty path and reads
`endpointName` from global configuration.

Applications that need explicit config directories, service names, queue sizes,
or injected runtime components must coordinate through globals or pre-start
configuration side effects.

### 5. Application state is not explicit enough

`Application` is a singleton and reconstructs runtime components after shutdown.
The current API does not explicitly expose or enforce lifecycle state.

The framework should make repeated startup, run-before-startup, shutdown
idempotency, and failed startup behavior predictable.

### 6. Application-level foreground services need a generic adapter

Some applications have a foreground local lifecycle, for example a CLI/TUI
process that still wants to reuse framework startup, EventHub, BackgroundRoutine,
and module infrastructure.

These foreground flows often naturally expose a simple lifecycle shape:

```go
type LifecycleService interface {
    Startup(ctx context.Context) error
    Run(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

The framework currently has no generic adapter from that shape into
`framework/service.Service`.

## Design Goals

- Preserve existing public APIs unless a new API can be added alongside them.
- Keep `Application` as the process runtime owner.
- Keep `Service` as the framework lifecycle coordinator.
- Keep `initiator` and `module` as plugin groups with distinct responsibilities.
- Make startup failures deterministic and observable.
- Make invalid plugin registration visible.
- Add compile-time-friendly interfaces while preserving reflection compatibility.
- Avoid moving application-specific business logic into `magicCommon/framework`.

## Non-Goals

- Do not add WorkOrch-specific concepts to magicCommon.
- Do not make `framework/service` depend on business modules.
- Do not replace EventHub, BackgroundRoutine, configuration, or health subsystems.
- Do not remove reflection-based plugin compatibility in the first pass.
- Do not change existing `Service` interface signatures in a breaking way.

## Proposed Improvements

### Phase 1: Plugin Registration Safety

Add error-returning registration functions:

```go
package initiator

func RegisterE(initiator any) error
func MustRegister(initiator any)
```

```go
package module

func RegisterE(module any) error
func MustRegister(module any)
```

Current `Register` can remain for compatibility, but it should log registration
errors at minimum.

Recommended behavior:

- `RegisterE`: returns the underlying validation or duplicate ID error.
- `MustRegister`: panics on registration error, intended for `init()` side-effect
  registration where failure must be fatal during development.
- `Register`: keeps existing signature and delegates to `RegisterE`; logs errors.

Acceptance:

- duplicate ID returns an error from `RegisterE`
- non-pointer plugin returns an error
- nil plugin returns an error
- missing `ID` or `Run` returns an error
- existing side-effect registration still compiles

### Phase 2: Explicit Plugin Interfaces

Add explicit interfaces in `framework/plugin/common`.

```go
type Plugin interface {
    ID() string
    Run(context.Context) *cd.Error
}

type Weighted interface {
    Weight() int
}

type Setupper interface {
    Setup(context.Context, event.Hub, task.BackgroundRoutine) *cd.Error
}

type Teardowner interface {
    Teardown(context.Context)
}
```

The manager may still support reflection for compatibility, but new validation
and invocation should prefer interface assertions.

Acceptance:

- plugins implementing explicit interfaces run without reflection for those methods
- legacy plugins discovered by reflection still work
- tests cover mixed explicit and legacy plugins
- method signature mismatch returns a clear validation or invocation error

### Phase 3: Manager-Level Setup Rollback

Enhance `PluginMgr.Setup` to record successful setup calls. If a later plugin
setup fails, rollback only the already-setup plugins in reverse order.

Suggested behavior:

- Treat missing `Setup` as no-op success.
- If plugin N setup fails, call `Teardown` on N-1..1.
- Return the original setup error.
- Log teardown errors without replacing the original setup error.

Acceptance:

- setup failure at plugin N tears down previously setup plugins in reverse order
- plugin N is not torn down unless its setup completed
- teardown errors are logged and do not mask setup failure
- `DefaultService` remains safe if it also calls group teardown after setup failure

### Phase 4: Application Startup Options

Add option-based application startup without breaking existing calls.

```go
type Options struct {
    ConfigDir string
    ServiceName string
    EventHubQueueSize int
    BackgroundQueueSize int
    EventHub event.Hub
    BackgroundRoutine task.BackgroundRoutine
    Ownership RuntimeOwnership
}

func StartupWithOptions(ctx context.Context, service service.Service, opts Options) *cd.Error
func NewApplication(opts Options) Application
```

Default behavior remains:

- config dir empty
- service name from `endpointName`, fallback `magicFramework`
- queue sizes from current defaults/env
- internally constructed EventHub and BackgroundRoutine

Acceptance:

- existing `application.Startup(ctx, service)` behavior is unchanged
- `StartupWithOptions` uses explicit config dir
- explicit service name overrides configuration service name
- injected EventHub/BackgroundRoutine are used when supplied
- shutdown still terminates owned runtime components correctly

Ownership rule:

- If EventHub/BackgroundRoutine are injected, ownership must be explicit.
- Default recommendation: Application should not terminate externally owned
  injected components unless Options declares ownership.

If ownership is added, prefer:

```go
type RuntimeOwnership struct {
    EventHub bool
    BackgroundRoutine bool
}
```

### Phase 5: Application State Guardrails

Add internal lifecycle states:

```go
type State string

const (
    StateNew State = "new"
    StateStarting State = "starting"
    StateRunning State = "running"
    StateFailed State = "failed"
    StateShutdown State = "shutdown"
)
```

State rules:

- `Run` before successful `Startup` returns a clear error.
- repeated `Startup` returns a clear error unless the previous lifecycle has shut down.
- `Shutdown` is idempotent.
- failed startup transitions to failed and performs best-effort cleanup.
- successful shutdown recreates default runtime components for the singleton path.

Acceptance:

- tests cover run-before-startup
- tests cover repeated startup
- tests cover shutdown before startup
- tests cover failed startup cleanup
- health manager state remains consistent with application state

### Phase 6: Lifecycle Service Adapter

Add a small adapter package or type that lets applications express local
foreground flows without implementing the full framework `Service` signature.

Proposed type:

```go
type LifecycleService interface {
    Startup(ctx context.Context) error
    Run(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

Adapter:

```go
func AdaptLifecycle(name string, svc LifecycleService) service.Service
```

Adapter behavior:

- `Startup(ctx, serviceName, eventHub, backgroundRoutine)` calls `svc.Startup(ctx)`.
- `Run(ctx)` calls `svc.Run(ctx)`.
- `Shutdown(ctx)` calls `svc.Shutdown(ctx)`.
- Adapter converts returned `error` to `*cd.Error`.
- Adapter name is used only as fallback metadata; the framework-provided
  service name remains authoritative.

Acceptance:

- adapter can run through `application.Startup`, `application.Run`,
  `application.Shutdown`
- startup/run errors are converted to `*cd.Error`
- shutdown is called on application shutdown
- adapter does not own EventHub or BackgroundRoutine unless the wrapped service
  receives them through an application-specific constructor

## Suggested Implementation Order

1. Add registration error APIs and tests.
2. Add explicit plugin interfaces and keep reflection compatibility.
3. Add setup rollback tests and implementation.
4. Add application state tests and guardrails.
5. Add startup options with backward-compatible defaults.
6. Add lifecycle adapter and documentation.

This order reduces risk because plugin registration and rollback are local to
the plugin manager, while Application options and lifecycle adapter touch broader
runtime semantics.

## Compatibility Strategy

The first implementation pass should be additive:

- keep existing `application.Startup`
- keep existing `service.Service`
- keep existing `initiator.Register`
- keep existing `module.Register`
- keep reflection fallback
- keep existing default singleton behavior

New projects should prefer:

- `RegisterE` or `MustRegister`
- explicit plugin interfaces
- `StartupWithOptions` when explicit runtime configuration is needed
- `AdaptLifecycle` for local foreground lifecycle flows

## Validation Plan

Run focused framework tests:

```bash
GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor \
go test ./framework/plugin/common ./framework/plugin/initiator ./framework/plugin/module ./framework/service ./framework/application -count=1
```

Run runtime lifecycle tests if Application/EventHub/BackgroundRoutine ownership
changes:

```bash
GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor \
go test ./event ./task ./framework/application -count=1
```

Run full project tests before final acceptance:

```bash
GOCACHE=/tmp/magiccommon-gocache GOFLAGS=-mod=vendor \
go test ./... -count=1
```

## Acceptance Checklist

- [x] Registration errors are observable.
- [x] Existing side-effect plugin registration still works.
- [x] Duplicate plugin IDs are tested.
- [x] Invalid plugin shapes are tested.
- [x] Setup rollback order is tested.
- [x] Teardown reverse order remains tested.
- [x] Existing Application startup path remains compatible.
- [x] Option-based startup is tested.
- [x] Application repeated startup / run-before-startup / idempotent shutdown are tested.
- [x] Lifecycle adapter is tested through real Application startup/run/shutdown.
- [x] Documentation and related skills are updated after implementation.

## Downstream Guidance

Application repositories should not place application code under
`vendor/github.com/muidea/magicCommon/framework`.

Application repositories should consume framework lifecycle APIs from their own
entry or application packages. If they need a simpler foreground lifecycle shape,
they should use the framework lifecycle adapter once it is available, or keep a
local adapter until magicCommon exposes one.

Framework modules should remain generic. Business concepts such as workspace,
run, approval, MCP capability, profile, or UI mode belong to application
repositories, not to magicCommon.
