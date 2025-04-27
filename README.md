# ✨ toJot

> *Your thoughts, beautifully organized.*

A simple, elegant note-taking application built with Wails, Vue 3, and TypeScript that helps you capture ideas when inspiration strikes!

## 🚀 About

Jot is a cross-platform desktop application designed to make note-taking a joy. With its clean interface and powerful features, Jot helps you organize your thoughts, ideas, and tasks all in one place. It features a modern rich text editor powered by TipTap, reliable local storage with Dexie.js, and seamless automatic updates.

## ✅ Features

- 📝 **Rich text editor** with support for formatting, lists, and tasks
- 💾 **Local database storage** keeps your notes safe and accessible
- 🔍 **Lightning-fast search** to find notes instantly
- 🎨 **Beautiful UI** built with Vue 3, Tailwind CSS, and DaisyUI
- 🖥️ **Cross-platform** compatibility (Windows, macOS)
- 🔄 **Automatic updates** so you're always on the latest version

## 👨‍💻 Development

### Prerequisites

- Go 1.18 or later
- Node.js 16 or later
- npm

### Live Development

To run in live development mode, run `wails dev` in the project directory. This launches a Vite development server with hot reloading for a smooth development experience. You can also access Go methods through the dev server at http://localhost:34115.

### Building

Ready to create a distributable version? Use:

```bash
wails build
```

For platform-specific builds, we've included handy scripts:
- Windows: `build.bat`
- macOS/Linux: `build.sh`

## 📁 Project Structure

- `frontend/`: Vue 3 + TypeScript frontend application
- `app.go`: Main application logic
- `updater.go`: Update service for checking and applying updates
- `main.go`: Entry point for the Wails application

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues or pull requests.
