# ELO Tracking System Tests

This document describes the comprehensive test suite for the Agent ELO tracking system.

## Test Files

### `elo_test.go` - Core ELO Logic Tests
Unit tests for the core ELO calculation and database logic.

**Key Test Cases:**
- `TestELOCalculator_CalculateELO`: Tests ELO rating calculations with various scenarios
  - Equal ratings with win/loss/draw outcomes
  - Underdog vs favorite matchups
  - Different K-factor impacts
- `TestELOCalculator_DetermineKFactor`: Tests dynamic K-factor assignment
  - New players (K=64), experienced players (K=32), masters (K=16)
- `TestELOCalculator_DetermineWinner`: Tests winner determination from task execution status
- `TestELOCalculator_RecordCompetition`: Tests full competition recording workflow
- `TestELOCalculator_ProcessTaskCompetitions`: Tests batch processing of task competitions
- `TestELOCalculator_PreventDuplicateCompetitions`: Tests duplicate prevention

### `elo_api_test.go` - API Endpoint Tests
Integration tests for all ELO and competition API endpoints.

**ELO API Tests:**
- `TestELOAPI_GetLeaderboard`: Tests `/api/elo/leaderboard` endpoint
- `TestELOAPI_GetAgentHistory`: Tests `/api/elo/agent/{id}/history` endpoint
- `TestELOAPI_GetHeadToHeadRecord`: Tests `/api/elo/head-to-head/{id1}/{id2}` endpoint

**Competitions API Tests:**
- `TestCompetitionsAPI_ListCompetitions`: Tests `/api/competitions` GET
- `TestCompetitionsAPI_GetCompetition`: Tests `/api/competitions/{id}` GET
- `TestCompetitionsAPI_CreateCompetition`: Tests `/api/competitions` POST
- `TestCompetitionsAPI_ProcessTask`: Tests `/api/competitions/process-task/{id}` POST
- `TestCompetitionsAPI_GetCompetitionHistory`: Tests `/api/competitions/history`
- `TestCompetitionsAPI_InvalidRequests`: Tests error handling for invalid requests

### `elo_integration_test.go` - End-to-End Integration Tests
Full workflow integration tests including automatic competition processing.

**Integration Tests:**
- `TestAutomaticCompetitionProcessing`: Tests automatic competition creation when task executions complete
  - Verifies competitions are created when 2+ agents finish tasks
  - Tests proper ELO rating updates
  - Validates win/loss/draw statistics
- `TestELORatingProgression`: Tests ELO rating changes over multiple competitions
  - Verifies proper rating progression with wins/losses/draws
  - Tests game count and statistics tracking
- `TestConcurrentCompetitionProcessing`: Tests concurrent status updates don't create duplicate competitions

## Test Utilities

### Database Setup
- `setupELOTestDB(t)`: Cleans all test data including ELO tables
- `createTestAgentsForELO(t, ctx)`: Creates test agents with proper foreign key setup
- `createTestCompetitionData(t, ctx)`: Creates full test dataset for API tests

### Test Data Management
All tests use isolated test databases with automatic cleanup:
- Unique database files per test run (`remote-code-test-{timestamp}.db`)
- Proper foreign key constraint handling in cleanup order
- Complete migration application including ELO schema

## Running Tests

### Run All ELO Tests
```bash
go test -v -run "ELO|Competition" -timeout 30s
```

### Run Specific Test Categories
```bash
# Core logic tests only
go test -v -run "TestELOCalculator"

# API tests only  
go test -v -run "TestELOAPI|TestCompetitionsAPI"

# Integration tests only
go test -v -run "TestAutomaticCompetitionProcessing|TestELORatingProgression"
```

### Run Individual Tests
```bash
# Test specific functionality
go test -v -run "TestELOCalculator_CalculateELO"
go test -v -run "TestCompetitionsAPI_CreateCompetition"
```

## Test Coverage

The test suite covers:

✅ **ELO Calculation Logic**
- Standard ELO formula implementation
- Dynamic K-factor assignment
- Win/loss/draw result handling

✅ **Database Operations**
- Competition recording
- Agent rating updates
- Duplicate prevention
- Foreign key relationships

✅ **API Endpoints**
- All GET endpoints with valid/invalid data
- POST endpoints with success/error scenarios
- JSON serialization/deserialization
- Error response handling

✅ **Automatic Processing**
- Async competition processing
- Task execution status monitoring
- Multiple agent competitions
- Concurrent update handling

✅ **Data Integrity**
- Proper transaction handling
- Foreign key constraint enforcement
- Statistical accuracy (games, wins, losses, draws)
- ELO rating conservation (rating changes sum to zero)

## Performance Notes

### Test Timing
- Unit tests: ~0.01s each
- API tests: ~0.01s each  
- Integration tests: ~1-2s each (due to async processing delays)

### Database Performance
- SQLite locking may cause occasional failures in concurrent tests
- Tests include appropriate delays for async processing
- Uses separate database files to prevent test interference

### Memory Usage
- Tests are designed to be memory-efficient
- Proper cleanup prevents memory leaks
- Isolated test data prevents cumulative memory usage

## Troubleshooting

### Common Issues

**Database Lock Errors:**
```
database is locked (5) (SQLITE_BUSY)
```
- This is expected in concurrent tests
- Tests are designed to handle this gracefully
- Indicates proper concurrent access protection

**Timing-Related Failures:**
- Integration tests may occasionally fail due to async processing timing
- Increase sleep durations in integration tests if needed
- Tests include generous timeouts for CI environments

**Migration Failures:**
```
failed to execute migration: table already exists
```
- Ensure proper test cleanup between runs
- Check that test database files are being removed
- Verify migration files are in correct location

### Debug Tips

1. **Enable Verbose Logging:**
   ```bash
   go test -v -run "TestName"
   ```

2. **Check Database State:**
   ```bash
   # Inspect test database after failure
   sqlite3 remote-code-test-*.db ".schema"
   sqlite3 remote-code-test-*.db "SELECT * FROM agent_competitions;"
   ```

3. **Timing Issues:**
   - Increase sleep durations in integration tests
   - Add logging to async processing functions
   - Use race detector: `go test -race`

## Contributing

When adding new ELO-related functionality:

1. Add unit tests to `elo_test.go`
2. Add API tests to `elo_api_test.go` 
3. Add integration tests to `elo_integration_test.go`
4. Update this documentation
5. Ensure all existing tests still pass
6. Verify test database cleanup is proper