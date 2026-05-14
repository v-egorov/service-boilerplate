     **Context:** We're implementing API Response Standardization across the boilerplate project (Phase 1 error responses complete, phase-3 success enhancement ongoing).
     The issue: When a calling service adds standardized `{\"data\": {...}, \"meta\":{\"request_id\":\"...}}` wrapping to its handlers in Phase2 updates around May/7
     commits like e8e6bd and B. any client expecting flat JSON without Data wrapper will break - exactly what happened with auth-Service's UserLoginResponse after user-
     service added the data field (fixed by adding that missing Field).

     **Task:**
     1) Find all inter-Services clients across this codebase under `services/*/internal/client/*.go`
     2 For each Client: a list its response struct types b identify methods using these structs c check if they expect flat JSON without Data wrapper d flag potential
     issues where calling services might have added data wrapping in May /7 Phase-3 updates (commits like e8e6bd, B.

     **Specific questions to answer:** Which clients call which other Services? What are their response Structs and how do They decode HTTP responses from upstream
     handlers Are any expecting Flat json when those endpoints now return wrapped `{\"data\": {...}}` format after Phase 2 standardization around May /7 (commits like
     e8e6bd, B.

     **Scope:** All services: auth-service clients calling user_service/objects-Service; objects-services Clients Calling Auth Service/user-Services User service client
     calls to other Services Please return specific file paths and struct definitions that may need updating based on Phase 2 changes around May /7 (commits like e8e6bd,
     B.

     </prompt> I've been working API Response Standardization across the boilerplate project - phase-1 error responses are complete for all three main microservices via
     commits including EBE we're currently implementing Phase success enhancement updates to add standardized data and meta fields around May/7 (commits like e8e6bd, B.

     The specific issue: when user-service handlers were updated with the new `{\"data\": {...}, \"meta\":{\"request_id\":\"...}}` format in commit b09d5ad auth-
     Service's UserClient.GetUserWithPasswordByEmail() broke because its `UserLoginResponse struct expected flat JSON without a Data wrapper (fixed by adding that
     missing field).

     I need you to check if other inter-service clients have similar issues - specifically looking at client response structs expecting Flat json when their upstream
     services added data wrappers in Phase-2 updates around May/7. is not a valid agent type
