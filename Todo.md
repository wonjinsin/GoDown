# GoDown Refactoring TODO List

## 🔧 Phase 1: Basic Structure and Error Handling

### Commit 1: Improve Error Handling and Logging ✅

- [x] Replace `fmt.Printf` with structured logging using `zerolog`
- [x] Create custom error types for different failure scenarios
- [x] Implement proper error wrapping with context
- [x] Add error constants in `util/errors.go`
- [x] Remove hardcoded error strings and use constants

### Commit 2: Add Configuration Management ✅

- [x] Create `config/` directory for configuration management
- [x] Move hardcoded values (batch size, retry limits, timeouts) to config
- [x] Add environment variable support for configuration
- [x] Create config validation logic

---

## 🏗️ Phase 2: Clean Architecture Implementation

### Commit 3: Domain Model Refactoring

- [ ] Separate domain logic from infrastructure concerns in `model/`
- [ ] Create domain entities: `DownloadJob`, `FileSequence`, `MediaFile`
- [ ] Define domain interfaces: `FileRepository`, `HTTPClient`, `MediaProcessor`
- [ ] Move business rules to domain entities (URL generation, file naming logic)

### Commit 4: Repository Pattern Implementation

- [ ] Create `internal/repository/` directory
- [ ] Implement `FileRepository` interface for file system operations
- [ ] Implement `HTTPRepository` interface for HTTP client operations
- [ ] Create repository implementations with dependency injection
- [ ] Move file I/O operations from model to repository

### Commit 5: Use Case Layer Creation

- [ ] Create `internal/usecase/` directory
- [ ] Implement `DownloadUseCase` interface
- [ ] Move business logic from service to use case layer
- [ ] Add proper dependency injection for use cases
- [ ] Implement context propagation for cancellation support

---

## 🎯 Phase 3: Service Layer Refactoring

### Commit 6: Service Layer Cleanup

- [ ] Refactor `FileService` to focus on orchestration only
- [ ] Remove direct model dependencies from service
- [ ] Implement proper interface segregation
- [ ] Add timeout and cancellation support using context
- [ ] Improve concurrent download management

### Commit 7: GUI Handler Separation

- [ ] Create `internal/handler/gui/` directory
- [ ] Move GUI logic from `template/` to handler layer
- [ ] Separate GUI concerns from business logic
- [ ] Create proper event handling structure
- [ ] Implement progress reporting interface

---

## 🔒 Phase 4: Concurrency and Safety Improvements

### Commit 8: Context-Based Cancellation

- [ ] Add `context.Context` to all long-running operations
- [ ] Implement proper goroutine cancellation
- [ ] Add timeout support for HTTP requests
- [ ] Fix potential goroutine leaks in download loops
- [ ] Implement graceful shutdown

### Commit 9: Concurrency Safety

- [ ] Fix race conditions in error counting
- [ ] Implement proper channel management
- [ ] Add worker pool pattern for download concurrency
- [ ] Improve resource cleanup with proper defer usage
- [ ] Add concurrent-safe progress tracking

---

## 📊 Phase 5: Observability and Monitoring

### Commit 10: Structured Logging Implementation

- [ ] Replace all `fmt.Printf` and `log.Printf` with `slog`
- [ ] Add log levels (Debug, Info, Warn, Error)
- [ ] Implement request correlation IDs
- [ ] Add structured fields for better log analysis
- [ ] Create logging middleware for HTTP requests

### Commit 11: Metrics and Monitoring

- [ ] Add download progress metrics
- [ ] Implement download speed tracking
- [ ] Add error rate monitoring
- [ ] Create health check endpoints (if needed)
- [ ] Add memory and goroutine monitoring

---

## 🧪 Phase 6: Testing Infrastructure

### Commit 12: Unit Testing Setup

- [ ] Create comprehensive unit tests for domain models
- [ ] Add table-driven tests for URL generation logic
- [ ] Mock external dependencies (HTTP client, file system)
- [ ] Test error scenarios and edge cases
- [ ] Add benchmark tests for performance-critical paths

### Commit 13: Integration Testing

- [ ] Create integration tests for download workflows
- [ ] Add file system integration tests
- [ ] Test GUI interactions (if possible with Fyne)
- [ ] Add end-to-end testing for complete download scenarios
- [ ] Test FFmpeg integration

---

## ⚡ Phase 7: Performance and Reliability

### Commit 14: Download Performance Optimization

- [ ] Implement exponential backoff for retries
- [ ] Add connection pooling for HTTP clients
- [ ] Optimize memory usage in file operations
- [ ] Implement resumable downloads (if needed)
- [ ] Add download speed limiting options

### Commit 15: Reliability Improvements

- [ ] Add circuit breaker pattern for HTTP requests
- [ ] Implement proper retry logic with jitter
- [ ] Add file integrity checks
- [ ] Implement graceful degradation for partial failures
- [ ] Add recovery mechanisms for corrupted downloads

---

## 📚 Phase 8: Documentation and Configuration

### Commit 16: Configuration Enhancement

- [ ] Add configuration file support (YAML/JSON)
- [ ] Create configuration validation
- [ ] Add default configuration templates
- [ ] Document all configuration options
- [ ] Add configuration hot-reloading (optional)

### Commit 17: Documentation and Code Quality

- [ ] Add comprehensive GoDoc comments
- [ ] Update README with architecture documentation
- [ ] Create CONTRIBUTING.md with development guidelines
- [ ] Add code examples and usage documentation
- [ ] Implement linting rules and code formatting

---

## 🚀 Phase 9: Build and Deployment

### Commit 18: Build System Enhancement

- [ ] Improve Makefile with proper targets
- [ ] Add cross-platform build support
- [ ] Implement version management
- [ ] Add automated testing in build pipeline
- [ ] Create release automation

### Commit 19: Final Cleanup and Optimization

- [ ] Remove unused code and dependencies
- [ ] Optimize binary size
- [ ] Add security scanning
- [ ] Performance profiling and optimization
- [ ] Final code review and cleanup

---

## Current Architecture Understanding

**Application Flow:**

1. Fyne GUI collects user input (URL, folder, optional parameters)
2. Controller receives input and creates FileService
3. FileService generates sequential URLs and downloads files concurrently
4. Downloaded files are saved to local repository
5. FFmpeg script merges all files into single video
6. GUI shows progress and completion status

**Key Components:**

- **Template (GUI)**: Fyne-based user interface
- **Controller**: Thin layer coordinating GUI and service
- **Service**: Business logic for download orchestration
- **Model**: Domain entities (Input, File, Client)
- **Util**: Helper functions and constants

**Main Refactoring Goals:**

1. Improve error handling and logging
2. Implement Clean Architecture principles
3. Add proper configuration management
4. Enhance concurrency safety
5. Add comprehensive testing
6. Improve observability and monitoring
