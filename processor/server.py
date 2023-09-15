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
OPENAI_API_KEY = "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"

os.environ['OPENAI_API_KEY'] = OPENAI_API_KEY
openai.api_key = os.getenv("OPENAI_API_KEY")

db_params = {
    "dbname": "curater",
    "user": "postgres",
    "password": "postgres",
    "host": "localhost",  # or the address of your PostgreSQL server
    "port": 5433  # Default PostgreSQL port
}

# db_params = {
#     "dbname": "curater",
#     "user": "postgres",
#     "password": "postgres",
#     "host": "192.168.3.25",  # or the address of your PostgreSQL server
#     "port": 5433  # Default PostgreSQL port
# }

# Initialize a global database connection object
db_conn = psycopg2.connect(**db_params)
db_cursor = db_conn.cursor()

class ResourceNotFound(Exception):
    """Custom exception class."""
    def __init__(self, message="resource not found"):
        self.message = message
        super().__init__(self.message)

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
        content = get_content(content_id)
        tags = get_all_tags()
        print(tags)
        summarWithTags = summarizeAndTag(content, tags)
        summary, tags = parse_summary_and_tags(summarWithTags)
        print("summary", summary)
        print("tags", tags)
        add_tag_to_content(content_id, tags, summary)

        resp = {
                "id": content_id,
                "summary": summary
            }
        return jsonify(resp)
    except ResourceNotFound as e:
        return jsonify({"error": "Content not found"}), 404
    except Exception as e:
        return jsonify({"error": str(e)}), 500


def add_tag_to_content(content_id, tags, summary):
    try: 
        db_cursor.execute("UPDATE content SET summary = %s WHERE id = %s RETURNING id;", (summary, content_id))

        tag_ids = []
        for tag in tags:
            db_cursor.execute("SELECT id FROM tags WHERE tag = %s;", (tag,))
            result = db_cursor.fetchone()
            if result:
                tag_ids.append(result[0])

        # Insert records into the content_tags table in a single query
        insert_query = """
        INSERT INTO content_tags (content_id, tag_id) VALUES 
        {}
        """.format(", ".join("(%s, %s)" % (content_id, tag_id) for tag_id in tag_ids))

        db_cursor.execute(insert_query)


        # Commit the transaction
        db_conn.commit()
        print("Tags inserted successfully!")
    except psycopg2.IntegrityError as e:
        print("error updating tags and summary !", e)
        raise e


def get_content(content_id): 
    # Query the "contents" table in your database
    db_cursor.execute("SELECT * FROM content WHERE id = %s;", (content_id,))
    result = db_cursor.fetchone()  # Assuming 'id' is unique, fetch one row
    if result:            
            return result[8]
    else:
            raise ResourceNotFound("content not found")
    

def get_all_tags():
    db_cursor.execute("SELECT tag FROM tags")

    # Fetch all the rows as a list of tuples
    tag_records = db_cursor.fetchall()

    # Extract the tag names from the tuples and store them in a list
    tags = [record[0] for record in tag_records]
    return tags

def summarizeAndTag(text, tags):
    # Define prompt
    prompt_template = """Write a concise summary of the newsletter strictly following below criteria. 
    1. Summary should only get key points and then group it into a paragraph with 30 to 40 words.
    2. Avoid introduction and decalrative sentences 
    3. Now associate 2 to 3 tags from  the following list which are relevant to this newsletter
    """+str(tags)+"""
    4. Output should strictly be of the format. Do not output in any other format other than below format.
       Summary:
       Tags:

    "{text}"
    CONCISE SUMMARY:"""
    print(prompt_template)
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
    print("summarizing...")
    return stuff_chain.run(docs)

def parse_summary_and_tags(input):
    # Split the input text into lines
    lines = input.split('\n')

    # Initialize variables to store the summary and tags
    summary = ""
    tags = []

    # Iterate through the lines to extract summary and tags
    for line in lines:
      if line.startswith("Summary: "):
         # Extract the summary by removing the "Summary: " prefix
         summary = line[len("Summary: "):]
      elif line.startswith("Tags: "):
         # Extract tags by splitting the comma-separated string into a list
         tags = line[len("Tags: "):].split(', ')

    # Print the parsed summary and tags
    return summary, tags




def get_text_chunks_langchain(text):
    text_splitter = CharacterTextSplitter(chunk_size=500, chunk_overlap=100)
    texts = text_splitter.split_text(text)
    docs = [Document(page_content=t) for t in texts]
    return docs

if __name__ == '__main__':
    print("Starting flask app")
    app.run(host='0.0.0.0', port=9223)