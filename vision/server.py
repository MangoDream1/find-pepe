from flask import Flask, request
import uuid
import os
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import layers
from keras.applications import resnet
from PIL import Image
import json

app = Flask(__name__)

FILE_UPLOAD_FOLDER = "./data"
ALLOWED_FILE_EXTENSIONS = [".gif", ".png", ".jpg"]
MODEL_WEIGHTS_LOCATION = "./models/pepe-model.h5"

image_size = (224, 224)
batch_size = 32

data_augmentation = keras.Sequential(
    [
        layers.RandomFlip("horizontal"),
        layers.RandomRotation(0.1),
        layers.RandomZoom(0.2),
    ]
)

if (os.path.isfile(MODEL_WEIGHTS_LOCATION) == False):
  raise Exception("Model weights not stored at %s; cannot load model" % MODEL_WEIGHTS_LOCATION)

def make_model(input_shape, num_classes):
  inputs = keras.Input(shape=input_shape)
  # Image augmentation block
  x = data_augmentation(inputs)
  x = resnet.preprocess_input(x)

  # Entry block
  x = layers.Rescaling(1.0 / 255)(x)
  x = layers.Conv2D(32, 3, strides=2, padding="same")(x)
  x = layers.BatchNormalization()(x)
  x = layers.Activation("relu")(x)

  x = layers.Conv2D(64, 3, padding="same")(x)
  x = layers.BatchNormalization()(x)
  x = layers.Activation("relu")(x)

  previous_block_activation = x  # Set aside residual

  for size in [128, 256, 512, 728]:
    x = layers.Activation("relu")(x)
    x = layers.SeparableConv2D(size, 3, padding="same")(x)
    x = layers.BatchNormalization()(x)

    x = layers.Activation("relu")(x)
    x = layers.SeparableConv2D(size, 3, padding="same")(x)
    x = layers.BatchNormalization()(x)

    x = layers.MaxPooling2D(3, strides=2, padding="same")(x)

    # Project residual
    residual = layers.Conv2D(size, 1, strides=2, padding="same")(
        previous_block_activation
    )
    x = layers.add([x, residual])  # Add back residual
    previous_block_activation = x  # Set aside next residual

  x = layers.SeparableConv2D(1024, 3, padding="same")(x)
  x = layers.BatchNormalization()(x)
  x = layers.Activation("relu")(x)

  x = layers.GlobalAveragePooling2D()(x)
  if num_classes == 2:
    activation = "sigmoid"
    units = 1
  else:
    activation = "softmax"
    units = num_classes

  x = layers.Dropout(0.5)(x)
  outputs = layers.Dense(units, activation=activation)(x)
  return keras.Model(inputs, outputs)

model = make_model(input_shape=image_size + (3,), num_classes=2)
model.load_weights(MODEL_WEIGHTS_LOCATION)

def split_file_name_and_extension(file_path):
  return os.path.splitext(file_path)

def rewrite_gif_to_png(file_path):
  new_file_path = split_file_name_and_extension(file_path)[0] + ".png"
      
  img = Image.open(file_path)
  img.save(new_file_path, optimize=True)
  img.close()

  os.remove(file_path)

  return new_file_path

@app.route('/health', methods = ["GET"])
def hello_world():
    return 'Hello World!'

@app.route("/", methods = ["POST"])
def predict():
  file = request.files['value']

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

  return json.dumps({ "score": round(float(score), 2) })

if __name__ == '__main__':
    app.run()