import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import Status from './pages/Status'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div style={{ padding: 32 }}>
      <h1>Cloudbridge Relay Admin UI</h1>
      <Status />
    </div>
  )
}

export default App
