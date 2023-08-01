FROM python:3.9-slim-buster

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY ./server.py ./server.py
COPY ./models ./models

EXPOSE 5000

ENV FLASK_APP=server.py

CMD [ "python3", "-m", "flask", "run", "--host=0.0.0.0"]