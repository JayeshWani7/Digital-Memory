#!/usr/bin/env python3
"""
Load sample test data into Digital Memory system.
This script generates sample Slack and GitHub events and posts them to the ingestion service.
"""

import json
import requests
import time
from datetime import datetime, timedelta
from typing import List, Dict, Any

# Configuration
INGESTION_SERVICE_URL = "http://localhost:8001"
SLACK_WEBHOOK = f"{INGESTION_SERVICE_URL}/webhook/slack"
GITHUB_WEBHOOK = f"{INGESTION_SERVICE_URL}/webhook/github"

# Sample Slack messages
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
    },
    {
        "text": "Storage decision: We're using S3 for media files instead of storing them in the database. This reduces DB load significantly.",
        "user": "david",
        "channel": "architecture",
        "ts": 1711099200.000001
    },
    {
        "text": "Switched to Kubernetes for orchestration. We get better resource management and auto-scaling out of the box. Deployment times dropped from 30min to 5min.",
        "user": "eve",
        "channel": "infrastructure",
        "ts": 1711099800.000001
    }
]

# Sample GitHub events
SAMPLE_GITHUB_EVENTS = [
    {
        "action": "opened",
        "number": 1234,
        "pull_request": {
            "id": 12345,
            "number": 1234,
            "title": "Implement database connection pooling",
            "body": "This PR implements connection pooling using pgbouncer for PostgreSQL. Reduces connection overhead by 60% and improves throughput.",
            "user": {
                "login": "alice",
                "id": 1
            },
            "state": "open",
            "html_url": "https://github.com/company/backend/pull/1234",
            "diff_url": "https://github.com/company/backend/pull/1234.diff"
        },
        "repository": {
            "id": 123,
            "name": "backend",
            "full_name": "company/backend",
            "html_url": "https://github.com/company/backend"
        },
        "sender": {
            "login": "alice",
            "id": 1
        }
    },
    {
        "action": "opened",
        "number": 1235,
        "pull_request": {
            "id": 12346,
            "number": 1235,
            "title": "Add API caching layer with Redis",
            "body": "Implements Redis caching for frequently accessed endpoints. Expected 40% reduction in database queries.",
            "user": {
                "login": "bob",
                "id": 2
            },
            "state": "open",
            "html_url": "https://github.com/company/backend/pull/1235",
            "diff_url": "https://github.com/company/backend/pull/1235.diff"
        },
        "repository": {
            "id": 123,
            "name": "backend",
            "full_name": "company/backend",
            "html_url": "https://github.com/company/backend"
        },
        "sender": {
            "login": "bob",
            "id": 2
        }
    },
    {
        "action": "opened",
        "number": 1236,
        "pull_request": {
            "id": 12347,
            "number": 1236,
            "title": "Migrate to event-driven architecture",
            "body": "This large refactor moves us from request-response to an event-driven model using Kafka. Enables better decoupling and scalability.",
            "user": {
                "login": "carol",
                "id": 3
            },
            "state": "open",
            "html_url": "https://github.com/company/backend/pull/1236",
            "diff_url": "https://github.com/company/backend/pull/1236.diff"
        },
        "repository": {
            "id": 123,
            "name": "backend",
            "full_name": "company/backend",
            "html_url": "https://github.com/company/backend"
        },
        "sender": {
            "login": "carol",
            "id": 3
        }
    }
]


def create_slack_event(message: Dict[str, Any]) -> Dict[str, Any]:
    """Convert a message to a Slack webhook event"""
    return {
        "token": "verification_token",
        "team_id": "T00000000",
        "api_app_id": "A00000000",
        "event": {
            "type": "message",
            "user": message["user"],
            "text": message["text"],
            "ts": message["ts"],
            "channel": message["channel"],
            "event_ts": message["ts"]
        },
        "type": "event_callback",
        "event_id": f"Ev{'%016x' % int(message['ts'])}",
        "event_time": int(message["ts"])
    }


def create_github_event(pr: Dict[str, Any]) -> Dict[str, Any]:
    """Convert a PR to a GitHub webhook event"""
    return {
        "action": pr["action"],
        "number": pr["number"],
        "pull_request": pr["pull_request"],
        "repository": pr["repository"],
        "sender": pr["sender"]
    }


def load_slack_messages():
    """Load sample Slack messages"""
    print("Loading Slack messages...")
    
    for message in SAMPLE_SLACK_MESSAGES:
        event = create_slack_event(message)
        
        try:
            response = requests.post(
                SLACK_WEBHOOK,
                json=event,
                timeout=5
            )
            
            if response.status_code == 200:
                print(f"✓ Loaded Slack message from {message['user']}")
            else:
                print(f"✗ Failed to load Slack message: {response.status_code}")
        except requests.exceptions.RequestException as e:
            print(f"✗ Error loading Slack message: {e}")
        
        time.sleep(0.5)  # Rate limit
    
    print(f"✓ Loaded {len(SAMPLE_SLACK_MESSAGES)} Slack messages\n")


def load_github_events():
    """Load sample GitHub events"""
    print("Loading GitHub events...")
    
    for pr in SAMPLE_GITHUB_EVENTS:
        event = create_github_event(pr)
        
        try:
            response = requests.post(
                GITHUB_WEBHOOK,
                json=event,
                timeout=5
            )
            
            if response.status_code == 200:
                print(f"✓ Loaded GitHub PR #{pr['number']} from {pr['sender']['login']}")
            else:
                print(f"✗ Failed to load GitHub event: {response.status_code}")
        except requests.exceptions.RequestException as e:
            print(f"✗ Error loading GitHub event: {e}")
        
        time.sleep(0.5)  # Rate limit
    
    print(f"✓ Loaded {len(SAMPLE_GITHUB_EVENTS)} GitHub events\n")


def verify_data_loaded():
    """Verify that data was loaded in the system"""
    print("Verifying data was loaded...")
    
    try:
        # Check ingestion service status
        response = requests.get(f"{INGESTION_SERVICE_URL}/status", timeout=5)
        if response.status_code == 200:
            status = response.json()
            print(f"✓ Ingestion service running")
            print(f"  - Request count: {status.get('request_count', 'N/A')}")
            print(f"  - Success count: {status.get('success_count', 'N/A')}")
            
            # Get queue stats
            if 'queue_stats' in status:
                print(f"  - Queue stats: {status['queue_stats']}")
        else:
            print(f"✗ Could not verify ingestion service")
    except requests.exceptions.RequestException as e:
        print(f"✗ Error verifying data: {e}")
    
    print("\nNote: AI processing is asynchronous. Allow 30-60 seconds for")
    print("knowledge extraction and embedding generation.")
    print("\nCheck progress with:")
    print("  psql -h localhost -U memory_user -d digital_memory")
    print("  SELECT processing_status, COUNT(*) FROM events GROUP BY processing_status;")
    print("  SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL;")


def check_services():
    """Check if services are running"""
    print("Checking if services are running...\n")
    
    services = [
        ("Ingestion Service", INGESTION_SERVICE_URL),
        ("API Service", "http://localhost:8000"),
        ("AI Service", "http://localhost:8002"),
    ]
    
    all_running = True
    for name, url in services:
        try:
            response = requests.get(f"{url}/health", timeout=2)
            if response.status_code == 200:
                print(f"✓ {name} is running")
            else:
                print(f"✗ {name} not responding correctly")
                all_running = False
        except requests.exceptions.ConnectionError:
            print(f"✗ {name} is not running (connection refused)")
            all_running = False
        except requests.exceptions.Timeout:
            print(f"✗ {name} is not responding (timeout)")
            all_running = False
    
    print()
    return all_running


def main():
    """Main entry point"""
    print("=" * 60)
    print("Digital Memory Layer - Sample Data Loader")
    print("=" * 60)
    print()
    
    # Check if services are running
    if not check_services():
        print("ERROR: Not all services are running.")
        print("\nStart services with:")
        print("  docker-compose up -d")
        print("\nOr run locally in separate terminals:")
        print("  cd backend/ingestion-service && go run cmd/main.go")
        print("  cd backend/api-service && go run cmd/main.go")
        print("  cd backend/ai-service && python -m app.main")
        return 1
    
    # Load test data
    load_slack_messages()
    load_github_events()
    
    # Verify
    verify_data_loaded()
    
    print("\n" + "=" * 60)
    print("Sample data loaded successfully!")
    print("=" * 60)
    print("\nNext steps:")
    print("1. Wait 30-60 seconds for AI processing")
    print("2. Test query with:")
    print('   curl -X POST http://localhost:8000/api/v1/query \\')
    print('     -H "Content-Type: application/json" \\')
    print('     -d \'{"query": "What database decisions were made?", "top_k": 5}\'')
    print("\nFor more examples, see docs/EXAMPLES.md")
    
    return 0


if __name__ == "__main__":
    exit(main())
