FROM python:3.9-slim-buster

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY ./ ./

EXPOSE 5000

ENV FLASK_APP=server.py

CMD ["waitress-serve", "--host=0.0.0.0", "--port=5000", "server:app"]