from langchain.text_splitter import CharacterTextSplitter
from langchain.chat_models import ChatOpenAI
from langchain.chains.llm import LLMChain

from langchain.schema import Document
from langchain.prompts.prompt import PromptTemplate

from langchain.chains.llm import LLMChain
from langchain.prompts import PromptTemplate
from langchain.chains.combine_documents.stuff import StuffDocumentsChain



from flask import Flask, request, jsonify
from langchain.chains.llm import LLMChain
import os
import openai

import psycopg2

app = Flask(__name__)
OPENAI_API_KEY = ""

os.environ['OPENAI_API_KEY'] = OPENAI_API_KEY
openai.api_key = os.getenv("OPENAI_API_KEY")

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
    return "we are doing good"


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

        print("parse data", content_id)
        # Query the "contents" table in your database
        db_cursor.execute("SELECT * FROM content WHERE id = %s;", (content_id,))
        result = db_cursor.fetchone()  # Assuming 'id' is unique, fetch one row
        
        if result:            
            summary = summarize(result[8])
            #todo save this
            resp = {
                "id": result[0],
                "content": result[1], 
                "summary": summary
            }
            return jsonify(resp)
        else:
            # users = [{'id': 1, 'username': 'Alice'}, {'id': 2, 'username': 'Bob'}]
            # return jsonify(users, status=200, mimetype='application/json')

            return jsonify({"error": "Content not found"}), 404
    except Exception as e:
        return "An error occurred: " + str(e)

def summarize(text):
    # Define prompt
    prompt_template = """Write a concise summary of the newsletter strictly following below criteria. 
    1. Summary should only get key points and then group it into a paragraph with 30 to 40 words.
    2. Avoid introduction and decalrative sentences 

    "{text}"
    CONCISE SUMMARY:"""
    prompt = PromptTemplate.from_template(prompt_template)

    docs = get_text_chunks_langchain(text)
    # Define LLM chain
    llm = ChatOpenAI(temperature=0, model_name="gpt-3.5-turbo-16k")
    llm_chain = LLMChain(llm=llm, prompt=prompt)

    # Define StuffDocumentsChain
    stuff_chain = StuffDocumentsChain(
        llm_chain=llm_chain, document_variable_name="text"
    )

    # docs = loader.load()
    print("summarizing")
    return stuff_chain.run(docs)

def get_text_chunks_langchain(text):
    text_splitter = CharacterTextSplitter(chunk_size=500, chunk_overlap=100)
    texts = text_splitter.split_text(text)
    docs = [Document(page_content=t) for t in texts]
    return docs

if __name__ == '__main__':
    print("Starting flask app")
    app.run(host='0.0.0.0', port=9223)