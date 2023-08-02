from flask import Flask, request
import uuid
import os
import tensorflow as tf
from tensorflow import keras
from PIL import Image
import json

app = Flask(__name__)

FILE_UPLOAD_FOLDER = "./data"
ALLOWED_FILE_EXTENSIONS = [".gif", ".png", ".jpg"]
MODEL_LOCATION = "./models/pepe-model.tf"

image_size = (224, 224)

if (os.path.isdir(MODEL_LOCATION) == False and os.path.isfile(MODEL_LOCATION) == False):
    raise Exception("Model not stored at %s; cannot load model" %
                    MODEL_LOCATION)


if (os.path.isdir(FILE_UPLOAD_FOLDER) == False):
    os.makedirs(FILE_UPLOAD_FOLDER)

model = tf.keras.models.load_model(MODEL_LOCATION)


def split_file_name_and_extension(file_path):
    return os.path.splitext(file_path)


def rewrite_gif_to_png(file_path):
    new_file_path = split_file_name_and_extension(file_path)[0] + ".png"

    img = Image.open(file_path)
    img.save(new_file_path, optimize=True)
    img.close()

    os.remove(file_path)

    return new_file_path


@app.route('/health', methods=["GET"])
def hello_world():
    return 'Hello World!'


@app.route("/", methods=["POST"])
def predict():
    file = request.files['file']

    file_extension = split_file_name_and_extension(file.filename)[1]
    if (file_extension not in ALLOWED_FILE_EXTENSIONS):
        return "Invalid file extension; got %s instead of allowed: [%s]" % (file_extension, ALLOWED_FILE_EXTENSIONS.join(", ")), 400

    id = str(uuid.uuid4())
    file_path = os.path.join(FILE_UPLOAD_FOLDER, id + file_extension)
    file.save(file_path)

    if (file_extension == ".gif"):
        file_path = rewrite_gif_to_png(file_path)

    img = keras.preprocessing.image.load_img(file_path, target_size=image_size)
    img_array = keras.preprocessing.image.img_to_array(img)
    img_array = tf.expand_dims(img_array, 0)

    os.remove(file_path)

    predictions = model.predict(img_array)
    score = predictions[0]

    return json.dumps({"score": round(float(score), 2)})


if __name__ == '__main__':
    app.run()
