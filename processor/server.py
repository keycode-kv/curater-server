from flask import Flask, render_template, request, jsonify
from exceptiongroup import catch

import psycopg2

app = Flask(__name__)

db_params = {
    "dbname": "curater",
    "user": "postgres",
    "password": "postgres",
    "host": "localhost",  # or the address of your PostgreSQL server
    "port": 5432  # Default PostgreSQL port
}

# Initialize a global database connection object
db_conn = psycopg2.connect(**db_params)
db_cursor = db_conn.cursor()


@app.route('/')
def index():
    return "i'm doing good"


@app.route('/health', methods=['GET'])
def health():
    return "health is ok"


@app.route('/summarize', methods=['POST'])
def summarize():
    try:
        request_data = request.get_json()
        if 'id' not in request_data:
            return jsonify({"error": "Missing 'id' in the request"}), 400

        # Extract the ID from the request JSON
        content_id = request_data['id']

        print("content_id", content_id)
        # Query the "contents" table in your database
        db_cursor.execute("SELECT * FROM content WHERE id = %s;", (content_id,))
        result = db_cursor.fetchone()  # Assuming 'id' is unique, fetch one row
        
        print("result", result)
        if result:
            print("result inside")
            # Convert the result to a dictionary and return as JSON
            content_dict = {
                "id": result[0],  # Replace with the actual column indices
                "other_column": result[1],  # Replace with the actual column names
                # Add more columns as needed
            }
            
            return jsonify(content_dict)
        else:
            print("result outside")
            # users = [{'id': 1, 'username': 'Alice'}, {'id': 2, 'username': 'Bob'}]
            # return jsonify(users, status=200, mimetype='application/json')

            return jsonify({"error": "Content not found"}), 404
    except Exception as e:
        return "An error occurred: " + str(e)


if __name__ == '__main__':
    print("Starting flask app")
    app.run(host='0.0.0.0', port=9223)