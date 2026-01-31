# Implementation Tasks: Advanced Spam Filtering

**Branch**: `009-advanced-spam-filtering` | **Spec**: [specs/009-advanced-spam-filtering/spec.md](spec.md)
**Status**: Planned

## Implementation Strategy

We will implement the feature in 3 distinct layers that build upon each other:
1.  **Persistence Layer**: Database tables for Greylist tuples and Bayes tokens.
2.  **Greylisting (P1)**: The "outer gate" that rejects connections early. 
3.  **Bayesian Filter (P2/P3)**: The "inner brain" that analyzes content and learns from user actions.

Each user story is designed to be independently testable. The foundational phase handles the shared database requirements.

## Phase 1: Setup & Contracts

- [x] T001 Define `SpamFilter`, `Greylister`, `BayesClassifier`, and `BayesTrainer` interfaces in `internal/core/ports/spam.go`.
- [x] T002 Define domain entities (`GreylistTuple`, `BayesToken`, `SpamScore`) in `internal/core/domain/spam.go`.
- [x] T003 Create directory structure: `internal/adapters/spam/greylist`, `internal/adapters/spam/bayesian`.

## Phase 2: Foundational (Persistance)

**Goal**: Establish storage for tokens and lists before building logic.

- [x] T004 Define SQLite schema for `greylist` table in `internal/adapters/storage/sqlite/migrations/00X_greylist.sql`.
- [x] T005 Define SQLite schema for `bayes_tokens` and `bayes_global` tables in `internal/adapters/storage/sqlite/migrations/00Y_bayes.sql`.
- [x] T006 Implement `GreylistRepository` in `internal/adapters/storage/sqlite/greylist_repo.go`.
- [x] T007 Implement `BayesRepository` in `internal/adapters/storage/sqlite/bayes_repo.go`.
- [x] T008 Register new repositories in `cmd/mailraven/serve.go`.

## Phase 3: User Story 1 - Greylisting (P1)

**Goal**: Block "fire-and-forget" bots by temporarily rejecting unknown sender/recipient pairs.

- [x] T009 [US1] Implement `GreylistService` logic (Check, Allow, Prune) in `internal/adapters/spam/greylist/service.go`.
- [x] T010 [US1] Update `ports.SpamFilter` implementation (e.g., `internal/adapters/spam/spam_filter.go`) to use `GreylistService`.
- [x] T011 [US1] Update `internal/adapters/smtp/server.go` to call `SpamFilter.CheckRecipient` in the `RCPT` command handler.
- [x] T012 [US1] Add configuration for `Greylist` (Enabled, RetryDelay, Expiration) in `internal/config/config.go`.
- [x] T013 [US1] [TEST] Create integration test in `tests/spam_greylist_test.go` verifying 451 response on first attempt and 250 on retry.

## Phase 4: User Story 2 - Bayesian Analysis (P2)

**Goal**: Score email content and route spam to Junk.

- [x] T014 [US2] Implement tokenizer (Unicode scanner) in `internal/adapters/spam/bayesian/tokenizer.go`.
- [x] T015 [US2] Implement `BayesClassifier` logic (Naive Bayes math) in `internal/adapters/spam/bayesian/classifier.go`.
- [x] T016 [US2] Implement `SpamCheckMiddleware` in `internal/adapters/smtp/middleware/spam.go` utilizing the classifier. (Integrated into `SpamProtectionService` and `server.go`)
- [x] T017 [US2] Register `SpamCheckMiddleware` in the SMTP server pipeline in `cmd/mailraven/serve.go`. (Via Service)
- [x] T018 [US2] Update `internal/adapters/imap/server.go` mechanism to verify `X-Spam-Status` header routing (if not already handled by delivery-time rules). (Handled in `smtp/handler.go`)
- [x] T019 [US2] [TEST] Unit test for Tokenizer and Classifier correctness in `internal/adapters/spam/bayesian/bayes_test.go`.

## Phase 5: User Story 3 - Feedback Loop (P3)

**Goal**: Enable users to train the filter by moving messages.

- [x] T020 [US3] Implement `BayesTrainer` logic (`TrainSpam`, `TrainHam`) in `internal/adapters/spam/bayesian/trainer.go`.
- [x] T021 [US3] Implement `COPY` command handler in `internal/adapters/imap/commands.go` (required for "Move").
- [x] T022 [US3] Add hook in `COPY` handler: If destination is `Junk`, call `TrainSpam`. If source is `Junk`, call `TrainHam`.
- [x] T023 [US3] [TEST] Integration test in `tests/spam_training_test.go`: Move message -> Verify token count increases in DB.

## Phase 6: Polish & Cross-Cutting

- [x] T024 Add background goroutine to prune expired Greylist entries in `internal/adapters/spam/greylist/pruner.go`.
- [x] T025 Add Prometheus metrics for `spam_detected_total`, `greylist_blocked_total` in `internal/observability/metrics.go`.
- [x] T026 Update `README.md` and `docs/COMPARISON_MOX.md` to reflect new capabilities.

## Dependencies

- **US3** depends on **US2** (Trainer needs Classifier structures) and **Foundational** (DB).
- **US1** depends on **Foundational** (DB).
- **US2** depends on **Foundational** (DB).

## Parallel Execution

- T006 (Greylist Repo) and T007 (Bayes Repo) are parallelizable.
- T009 (Greylist Logic) and T014 (Tokenizer) are parallelizable.
