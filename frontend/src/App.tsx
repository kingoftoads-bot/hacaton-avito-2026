import React, { useState } from 'react'

const AVITO_GREEN = '#97EB33'
const AVITO_BLUE = '#009CF0'
const AVITO_PURPLE = '#965EEB'

type ChatMessage = {
  role: 'user' | 'bot'
  text: string
}

function App() {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'bot', text: 'Привет! Я твой ИИ-наставник. Помогу с задачами, оформлю отгул или запрошу доступ.' }
  ])
  const [loading, setLoading] = useState(false)

  const sendMessage = async (text: string) => {
    if (!text.trim()) return
    setMessages(prev => [...prev, { role: 'user', text }])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch('http://localhost:8080/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: text, role: 'intern_backend' })
      })
      const data = await res.json()
      setMessages(prev => [...prev, { role: 'bot', text: data.reply || 'Ошибка: пустой ответ' }])
    } catch (e) {
      setMessages(prev => [...prev, { role: 'bot', text: 'Ошибка соединения с бэкендом.' }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ fontFamily: 'Arial, sans-serif', backgroundColor: '#F5F5F5', minHeight: '100vh', padding: '40px 20px', display: 'flex', justifyContent: 'center' }}>
      <div style={{ width: '100%', maxWidth: '800px' }}>
        
        {/* Шапка */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '20px' }}>
          <div style={{ display: 'flex' }}>
            <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_GREEN }}></div>
            <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_BLUE, marginLeft: '-10px' }}></div>
            <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_PURPLE, marginLeft: '-10px' }}></div>
          </div>
          <h1 style={{ margin: 0, fontSize: '24px' }}>AI-buddy</h1>
        </div>

        {/* Область чата */}
        <div style={{ background: '#fff', borderRadius: '12px', padding: '20px', minHeight: '400px', maxHeight: '500px', overflowY: 'auto', marginBottom: '20px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
          {messages.map((msg, i) => (
            <div key={i} style={{ marginBottom: '15px', display: 'flex', justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start' }}>
              <div style={{ 
                padding: '12px 16px', borderRadius: '12px', maxWidth: '80%', whiteSpace: 'pre-wrap',
                background: msg.role === 'user' ? AVITO_GREEN : '#F0F0F0',
                color: '#000'
              }}>
                {msg.text}
              </div>
            </div>
          ))}
          {loading && <div style={{ color: '#888', fontStyle: 'italic' }}>ИИ-наставник печатает...</div>}
        </div>

        {/* Кнопки быстрых действий */}
        <div style={{ display: 'flex', gap: '10px', marginBottom: '15px', flexWrap: 'wrap' }}>
          <button onClick={() => sendMessage('Какие у меня задачи на 1 неделю?')} style={{ padding: '8px 16px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer' }}>📋 Задачи на неделю</button>
          <button onClick={() => sendMessage('Хочу взять отгул на завтра')} style={{ padding: '8px 16px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer' }}>📅 Оформить отгул</button>
          <button onClick={() => sendMessage('Как получить доступ к Jira?')} style={{ padding: '8px 16px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer' }}>🔐 Запросить доступ</button>
        </div>

        {/* Поле ввода */}
        <div style={{ display: 'flex', gap: '10px', background: '#fff', padding: '10px', borderRadius: '12px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
          <input 
            value={input} 
            onChange={(e) => setInput(e.target.value)} 
            onKeyDown={(e) => e.key === 'Enter' && sendMessage(input)}
            placeholder="Напиши сообщение..."
            style={{ flex: 1, padding: '12px', border: 'none', outline: 'none', fontSize: '16px' }}
          />
          <button 
            onClick={() => sendMessage(input)} 
            disabled={loading}
            style={{ padding: '12px 24px', background: AVITO_GREEN, border: 'none', borderRadius: '8px', fontWeight: 'bold', cursor: 'pointer' }}
          >
            Отправить
          </button>
        </div>

      </div>
    </div>
  )
}

export default App