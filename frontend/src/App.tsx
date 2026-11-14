import React, { useState } from 'react'

// Цвета Авито (из презентации)
const AVITO_GREEN = '#97EB33' // Яркий салатовый
const AVITO_BLUE = '#009CF0'  // Голубой
const AVITO_PURPLE = '#965EEB' // Фиолетовый
const AVITO_RED = '#FF4053'   // Красный

function App() {
  const [message, setMessage] = useState('')
  const [reply, setReply] = useState('')
  const [loading, setLoading] = useState(false)

  const sendMessage = async () => {
     if (!message.trim()) return
    setLoading(true)
    setReply('')
    try {
      const res = await fetch('http://localhost:8080/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message, role: 'intern_backend' })
      })
      const data = await res.json()
      console.log("Ответ от бэкенда:", data) // <-- ВЫВЕДЕТ В КОНСОЛЬ
      
      if (data.reply) {
        setReply(data.reply)
      } else if (data.error) {
        setReply("ОШИБКА БЭКЕНДА: " + data.error)
      } else {
        setReply("Странный ответ: " + JSON.stringify(data))
      }
    } catch (e: any) {
      setReply('Ошибка сети: ' + e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ 
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      backgroundColor: '#F5F5F5', 
      minHeight: '100vh', 
      padding: '40px 20px',
      display: 'flex',
      justifyContent: 'center'
    }}>
      <div style={{ width: '100%', maxWidth: '700px' }}>
        
        {/* Шапка */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '10px' }}>
          <div style={{ width: '30px', height: '30px', borderRadius: '50%', background: AVITO_GREEN }}></div>
          <div style={{ width: '30px', height: '30px', borderRadius: '50%', background: AVITO_BLUE, marginLeft: '-15px' }}></div>
          <div style={{ width: '30px', height: '30px', borderRadius: '50%', background: AVITO_PURPLE, marginLeft: '-15px' }}></div>
          <h1 style={{ margin: 0, marginLeft: '10px', fontSize: '28px', color: '#000' }}>AI-buddy</h1>
        </div>
        <p style={{ color: '#666', marginBottom: '30px', fontSize: '16px' }}>
          Привет! Я твой ИИ-наставник. Помогу с задачами, отгулами и доступами.
        </p>

        {/* Поле ввода */}
        <div style={{ 
          display: 'flex', gap: '10px', marginBottom: '20px',
          background: '#fff', padding: '10px', borderRadius: '12px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.05)'
        }}>
          <input 
            value={message} 
            onChange={(e) => setMessage(e.target.value)} 
            onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
            placeholder="Например: Какие у меня задачи на 1 неделю?"
            style={{ 
              flex: 1, padding: '12px', border: '1px solid #E0E0E0', 
              borderRadius: '8px', fontSize: '16px', outline: 'none'
            }}
          />
          <button 
            onClick={sendMessage} 
            disabled={loading}
            style={{ 
              padding: '12px 24px', background: loading ? '#CCC' : AVITO_GREEN, 
              border: 'none', borderRadius: '8px', fontSize: '16px', 
              fontWeight: 'bold', cursor: loading ? 'wait' : 'pointer', color: '#000'
            }}
          >
            {loading ? 'Думаю...' : 'Отправить'}
          </button>
        </div>

        {/* Ответ */}
        {reply && (
          <div style={{ 
            padding: '20px', background: '#fff', borderRadius: '12px', 
            boxShadow: '0 2px 8px rgba(0,0,0,0.05)', borderLeft: `4px solid ${AVITO_GREEN}`
          }}>
            <strong style={{ display: 'block', marginBottom: '8px', color: '#000' }}>Ответ:</strong> 
            <span style={{ color: '#333', lineHeight: '1.5', whiteSpace: 'pre-wrap' }}>{reply}</span>
          </div>
        )}
      </div>
    </div>
  )
}

export default App