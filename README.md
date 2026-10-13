# Mreme / Meme Mirror

## Ejecutar en Raspberry Pi

Con Python 3.11 o superior instalado, ejecuta estos comandos desde una terminal
del escritorio de la Raspberry (Linux de 64 bits con `apt` y una webcam USB):

```bash
sudo apt-get update
sudo apt-get install -y git
git clone https://github.com/Renato-Roggia/Muestra2.git
cd Muestra2
bash install.sh
bash run.sh
```

Si ya descargaste este proyecto, abre una terminal en su carpeta y ejecuta solo:

```bash
bash install.sh
bash run.sh
```

`install.sh` instala las dependencias del sistema y las de Python en `.venv`.
Solicitará tu contraseña de `sudo` para los paquetes del sistema. No ejecutes
el script completo con `sudo`. No necesitas Docker ni Make.

Para volver a abrir la aplicación:

```bash
cd Muestra2
bash run.sh
```

Necesitas internet para instalar y para la primera descarga de modelos.
Después se reutilizan los modelos locales. La aplicación abre una ventana:
ejecútala desde el escritorio, no desde una sesión SSH sin pantalla.
Pulsa **Q** para salir. Si necesitas seleccionar otra cámara:
`CAMERA_INDEX=1 bash run.sh`.

A real-time pose detection application built with Python, OpenCV, and MediaPipe that displays specific monkey memes based on your body poses. The app uses computer vision to instantly detect and match four distinct states: Staring (neutral), Thinking (hand on mouth), Pointing (finger up), and Shocked (open mouth), displaying the corresponding monkey meme template in real-time.

<p align="center">
  <img src="memes/staring.jpg" width="180" />
  <img src="memes/thinking.jpg" width="180" />
  <img src="memes/pointing.jpg" width="180" />
  <img src="memes/shocked.jpg" width="180" />
</p>

## Features

- Real-time pose detection using MediaPipe Holistic
- Automatic monkey meme matching based on detected poses
- Multiple display modes (split view, side-by-side, picture-in-picture)
- Screenshot and video recording capabilities
- Face overlay onto monkey meme templates
- Vertical mode for social media export

## Supported Poses

| Pose | How to Trigger |
|------|----------------|
| Staring | Default state - just look at the camera |
| Thinking | Touch your finger to your chin |
| Pointing | Point your finger above your head |
| Shocked | Open your mouth wide |

## Requirements

- Python 3.11+
- Computer with a webcam
- Libraries in requirements.txt

The first launch downloads the MediaPipe pose, face, and hand models into `models/`.
An internet connection is needed for that first launch; later launches use the local files.

## Setup and running program

### Raspberry Pi: install and run

With the repository copied or downloaded onto your Raspberry Pi, open a terminal
in its folder and run:

```bash
bash install.sh
bash run.sh
```

The installer uses your existing Python 3.11+ installation, installs native
libraries through `apt` (requesting `sudo` when needed), and installs
`requirements.txt` into a local `.venv`. Run it as your normal user, without
`sudo`. Docker, Make, and manual virtual environment activation are not needed.
It requires 64-bit Linux with `apt`, internet access, and permission to use `sudo`.
Python packages are installed from prebuilt wheels, avoiding compilation on the Pi.

For subsequent launches, only run `bash run.sh`. Run `bash install.sh` again
after dependency changes; it reuses the environment and preserves your files.
Both scripts also work when invoked by their full path from another directory.
The first application launch may download the detection models into `models/`.

Troubleshooting:

- **No matching distribution:** check that Python and Linux are 64-bit and that
  your Python version has prebuilt packages for the dependencies. The installer
  intentionally stops instead of compiling them.
- **Existing unusable `.venv`:** move that folder aside and rerun the installer.
  A virtual environment copied from Windows cannot be reused on Linux.
- **No graphical session:** run `bash run.sh` in your desktop terminal; the app
  opens a window and cannot display it in a plain SSH session.
- **Camera cannot open:** connect the USB webcam and close other apps using it.
  To select a camera, run `CAMERA_INDEX=1 bash run.sh`.
- **Download or apt error:** check the connection and reported error, then rerun
  `bash install.sh`.

### Manual installation / other platforms

First, clone the repo
```bash
git clone https://github.com/Renato-Roggia/Muestra2.git
cd Muestra2
```

Install necessary dependencies
```bash
pip install -r requirements.txt
```

Use the same Python installation for setup and launch. For example, on Windows:
```powershell
py -3.11 -m pip install -r requirements.txt
py -3.11 main.py
```

Run program
```bash
python main.py
```

## Controls

| Key | Function |
|-----|----------|
| Q | Quit |
| H | Show/hide help overlay |
| S | Take screenshot |
| R | Start/stop video recording |
| G | Export GIF frames |
| F | Toggle fullscreen |
| P | Pause |
| M | Meme-only view |
| B | Side-by-side view |
| I | Picture-in-picture mode |
| V | Vertical mode (9:16 aspect ratio) |
| W | Toggle watermark |
| A | Toggle face overlay on meme |
| Z | Toggle face zoom |
| N | Toggle smooth transitions |
| C | Switch camera |
| ←/→ | Manually cycle through memes |
| ↑ | Return to auto-detection |
