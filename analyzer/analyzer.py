import time
import json
import subprocess
from elasticsearch import Elasticsearch
from elasticsearch.exceptions import NotFoundError
from datetime import datetime, timedelta
import logging

# Configure logging
logging.basicConfig(
    filename='../logs/analyzer.log',
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

# Initialize Elasticsearch client
es = Elasticsearch(['http://localhost:9200'])

# Load configuration from JSON file
def load_config(file_path):
    with open(file_path, 'r') as f:
        return json.load(f)

# Build query for logs
def build_query():
    now = datetime.utcnow()
    five_minutes_ago = now - timedelta(minutes=5)
    now_str = now.strftime('%Y-%m-%dT%H:%M:%S.%fZ')
    five_minutes_ago_str = five_minutes_ago.strftime('%Y-%m-%dT%H:%M:%S.%fZ')
    
    return {
        "query": {
            "bool": {
                "must": [
                    {"wildcard": {"message": "*error*"}},  # Case-insensitive search using wildcard
                    {
                        "range": {
                            "@timestamp": {
                                "gte": five_minutes_ago_str,  # Start of time range
                                "lte": now_str  # End of time range
                            }
                        }
                    }
                ],
                "must_not": [
                    {
                        "wildcard": {
                            "message": "*analyzer.log*"
                        }
                    },
                    {
                        "match_phrase": {
                            "log.file.path": "/logs/analyzer.log"
                        }
                    }
                ]
            }
        }
    }

# Function to query Elasticsearch
def get_logs_containing_error(index_pattern, query):
    try:
        response = es.search(index=index_pattern, body=query)
        hits = response['hits']['hits']
        logging.info(f"Found {len(hits)} log entries matching the query.")
        return hits
    except NotFoundError:
        logging.warning(f"No indices matching pattern {index_pattern} found.")
        return []
    except Exception as e:
        logging.error(f"An error occurred while querying Elasticsearch: {e}")
        return []

# Function to perform additional searches based on keyword matches
def perform_additional_search(application,search_pattern):
    additional_query = {
        "query": {
            "match": {
                "message": search_pattern
            }
        }
    }
    response = es.search(index='myindex-*', body=additional_query)
    hits = response['hits']['hits']
    logging.info(f"Performed additional search with pattern '{search_pattern}', found {len(hits)} entries.")
    return hits

# Function to run a command and check its output
def run_command(command):
    try:
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            logging.info(f"Command successful: {command}")
            return result.stdout
        else:
            logging.error(f"Command failed: {command}\nError: {result.stderr}")
            return None
    except Exception as e:
        logging.error(f"An error occurred while running command: {e}")
        return None

# Function to perform actions based on the configuration
def perform_actions(actions,healthcheckcommand):
    for action in actions:
        logging.info(f"Performing action: {action['action']}")
        result = run_command(action['action'])
        time.sleep(10)
        health_check_output = run_command(healthcheckcommand)
        if health_check_output:
            logging.info(f"Health check output: {health_check_output}")
            logging.info(f"Action '{action['action']}' completed successfully.")
            return
        else:
            logging.warning(f"Action '{action['action']}' did not complete as expected.")

# Main loop to run the query every minute
def main():
    config = load_config('config.json')
    keywords = config.get('keywords', {})

    index_pattern = 'myindex-*'
    
    while True:
        logging.info("Starting new iteration to query logs.")
        query = build_query()
        logs = get_logs_containing_error(index_pattern, query)
        
        for log in logs:
            timestamp = log['_source'].get('@timestamp')
            message = log['_source'].get('message')
            file_name = log['_source'].get('log', {}).get('file', {}).get('path', 'unknown')
            logging.info(f"Log entry: Timestamp: {timestamp}, Message: {message}, Log File: {file_name}")

            # Check for keywords and perform additional search
            for keyword, details in keywords.items():
                if keyword in message:
                    
                    logging.info(f"Keyword '{keyword}' found in log. Performing additional search.")
                    additional_logs = perform_additional_search(keyword,details['SuccessLog'])
                    if additional_logs:
                        
                        logging.info(f"Additional logs found for keyword '{keyword}'.")
                        # Run health check
                        health_check_output = run_command(details['healthcheck'])
                        if health_check_output:
                            logging.info(f"Health check output: {health_check_output}")
                            continue

                        # Perform actions
                        if 'Actions' in details:
                            perform_actions(details['Actions'],details['healthcheck'])

                        # Validate recovery (you might want to implement specific validation here)
                        logging.info(f"Validating recovery for keyword '{keyword}'.")
                        # Example: Check if service is healthy
                        validation_output = run_command(details['healthcheck'])
                        if validation_output:
                            logging.info(f"Service is healthy after recovery: {validation_output}")
                        else:
                            logging.warning(f"Service is not healthy after recovery.")
        
        logging.info("Waiting for 10 secs before the next iteration.")
        time.sleep(10)

if __name__ == "__main__":
    main()
