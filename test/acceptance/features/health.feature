Feature: Service Health
    As a support engineer,
    I want to get the current health status
    so that the service can be debugged.

Scenario: Service is live
    When a liveness request comes in
    Then a successful health response is returned

Scenario: Service is ready
    When a readiness request comes in
    Then a successful health response is returned
