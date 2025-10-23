import requests

# Use curl or requests library
def fetch_data():
    # curl alternative in Python
    response = requests.get('https://api.example.com')
    return response.json()

