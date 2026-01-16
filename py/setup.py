"""Setup script for multi-ws-tts-sdk"""

from setuptools import setup, find_packages
from pathlib import Path

# 读取 README
readme_file = Path(__file__).parent / "README.md"
long_description = readme_file.read_text(encoding="utf-8") if readme_file.exists() else ""

setup(
    name="multi-ws-tts-sdk",
    version="1.0.0",
    author="inccleo",
    description="Multi-Context WebSocket TTS SDK for Python",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/inccleo/multi-ws-tts-sdk",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Multimedia :: Sound/Audio :: Speech",
    ],
    python_requires=">=3.8",
    install_requires=[
        "websockets>=12.0",
    ],
    keywords="tts websocket text-to-speech multi-context",
    project_urls={
        "Bug Reports": "https://github.com/inccleo/multi-ws-tts-sdk/issues",
        "Source": "https://github.com/inccleo/multi-ws-tts-sdk",
    },
)
