import json

SAMPLE_SLACK_MESSAGES = [
    {
        "text": "We've decided to migrate our main database from MongoDB to PostgreSQL. This will give us better transaction support and ACID guarantees. Migration starts next sprint.",
        "user": "alice",
        "channel": "engineering",
        "ts": 1711097400.000001
    },
    {
        "text": "The API response time has been optimized by 40% after implementing caching with Redis. We're now getting sub-100ms responses on average.",
        "user": "bob",
        "channel": "performance",
        "ts": 1711098000.000001
    },
    {
        "text": "New microservice architecture is live! We've separated auth-service, user-service, and order-service. Each can scale independently now.",
        "user": "carol",
        "channel": "deployment",
        "ts": 1711098600.000001
    }
]

SAMPLE_GITHUB_EVENTS = [
    {
        "action": "opened",
        "number": 1234,
        "pull_request": {
            "title": "Implement database connection pooling",
            "body": "This PR implements connection pooling using pgbouncer for PostgreSQL. Reduces connection overhead by 60%."
        }
    },
    {
        "action": "opened",
        "number": 1235,
        "pull_request": {
            "title": "Add API caching layer with Redis",
            "body": "Implements Redis caching for frequently accessed endpoints."
        }
    }
]

if __name__ == "__main__":
    print("Sample data definitions")
    print(f"Slack messages: {len(SAMPLE_SLACK_MESSAGES)}")
    print(f"GitHub events: {len(SAMPLE_GITHUB_EVENTS)}")
