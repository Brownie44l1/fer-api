import requests
import numpy as np

# Generate random 48x48 image
random_image = np.random.rand(48 * 48).astype(np.float32).tolist()

response = requests.post(
    "http://localhost:8080/predict",
    json={"image": random_image}
)

print(response.json())