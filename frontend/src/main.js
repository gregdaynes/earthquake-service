import { register as registerApp } from './app.js'

const app = () => {
  registerApp()
}

document.addEventListener('DOMContentLoaded', app)
