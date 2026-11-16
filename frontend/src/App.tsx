import React, { useState, useEffect } from 'react'
import Markdown from 'react-markdown'

const AVITO_GREEN = '#97EB33'
const AVITO_BLUE = '#009CF0'
const AVITO_PURPLE = '#965EEB'

type ChatMessage = { role: 'user' | 'bot', text: string }
type Task = { id: number, title: string, description: string, is_done: boolean }

function App() {
  const [userName, setUserName] = useState('')
  const [userRole, setUserRole] = useState('intern_backend')
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(false)
  const [progress, setProgress] = useState({ total: 0, done: 0, percent: 0 })
  const [tasks, setTasks] = useState<Task[]>([])

  const loadData = () => {
    fetch(`http://localhost:8080/api/progress?role=${userRole}`)
      .then(res => res.json())
      .then(data => setProgress(data))

    fetch(`http://localhost:8080/api/tasks?role=${userRole}`)
      .then(res => res.json())
      .then(data => setTasks(data || []))
      .catch(() => {})
  }

  useEffect(() => {
    if (isLoggedIn) loadData()
  }, [userRole, isLoggedIn])

  const completeTask = async (id: number) => {
    await fetch(`http://localhost:8080/api/tasks/${id}/complete`, { method: 'POST' })
    await loadData()
  }

  const sendMessage = async (text: string) => {
    if (!text.trim()) return
    setMessages(prev => [...prev, { role: 'user', text }])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch('http://localhost:8080/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: text, role: userRole, name: userName })
      })
      const data = await res.json()
      setMessages(prev => [...prev, { role: 'bot', text: data.reply || 'Ошибка: пустой ответ' }])
    } catch (e) {
      setMessages(prev => [...prev, { role: 'bot', text: 'Ошибка соединения с бэкендом.' }])
    } finally {
      setLoading(false)
    }
  }

  if (!isLoggedIn) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#F5F5F5', fontFamily: 'Arial' }}>
        <div style={{ background: '#fff', padding: '40px', borderRadius: '16px', boxShadow: '0 4px 12px rgba(0,0,0,0.1)', width: '350px' }}>
          <div style={{ display: 'flex', gap: '10px', marginBottom: '20px', justifyContent: 'center' }}>
            <div style={{ width: '20px', height: '20px', borderRadius: '50%', background: AVITO_GREEN }}></div>
            <div style={{ width: '20px', height: '20px', borderRadius: '50%', background: AVITO_BLUE }}></div>
            <div style={{ width: '20px', height: '20px', borderRadius: '50%', background: AVITO_PURPLE }}></div>
          </div>
          <h2 style={{ textAlign: 'center', marginTop: 0 }}>Вход в AI-buddy</h2>
          <input 
            placeholder="Введи свое имя" 
            value={userName} 
            onChange={e => setUserName(e.target.value)}
            style={{ width: '100%', padding: '12px', marginBottom: '15px', borderRadius: '8px', border: '1px solid #ccc', boxSizing: 'border-box' }}
          />
          <select 
            value={userRole} 
            onChange={e => setUserRole(e.target.value)}
            style={{ width: '100%', padding: '12px', marginBottom: '20px', borderRadius: '8px', border: '1px solid #ccc', boxSizing: 'border-box' }}
          >
            <option value="intern_backend">Стажёр Backend</option>
            <option value="intern_frontend">Стажёр Frontend</option>
            <option value="mentor">Наставник</option>
          </select>
          <button 
            onClick={() => { if(userName.trim()) setIsLoggedIn(true); else alert('Введи имя!') }}
            style={{ width: '100%', padding: '12px', background: AVITO_GREEN, border: 'none', borderRadius: '8px', fontWeight: 'bold', cursor: 'pointer' }}
          >
            Войти
          </button>
        </div>
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'Arial, sans-serif', backgroundColor: '#F5F5F5', minHeight: '100vh', padding: '20px' }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto', display: 'grid', gridTemplateColumns: '1fr 350px', gap: '20px' }}>
        <div style={{ background: '#fff', borderRadius: '12px', padding: '20px', display: 'flex', flexDirection: 'column', height: '85vh', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '20px', borderBottom: '1px solid #eee', paddingBottom: '15px' }}>
            <div style={{ width: '30px', height: '30px', borderRadius: '50%', background: AVITO_GREEN }}></div>
            <h2 style={{ margin: 0 }}>AI-buddy</h2>
          </div>
          
          <div style={{ flex: 1, overflowY: 'auto', marginBottom: '20px' }}>
            {messages.map((msg, i) => (
              <div key={i} style={{ marginBottom: '15px', display: 'flex', justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start' }}>
                <div style={{ padding: '12px 16px', borderRadius: '12px', maxWidth: '80%', background: msg.role === 'user' ? AVITO_GREEN : '#F0F0F0', color: '#000' }}>
                  {msg.role === 'bot' ? <Markdown>{msg.text}</Markdown> : msg.text}
                </div>
              </div>
            ))}
            {loading && <div style={{ color: '#888', fontStyle: 'italic' }}>ИИ-наставник печатает...</div>}
          </div>

          <div style={{ display: 'flex', gap: '10px', marginBottom: '10px', flexWrap: 'wrap' }}>
            <button onClick={() => sendMessage('Какие у меня задачи?')} style={{ padding: '6px 12px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer', fontSize: '12px' }}>📋 Задачи</button>
            <button onClick={() => sendMessage('Хочу отгул')} style={{ padding: '6px 12px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer', fontSize: '12px' }}>📅 Отгул</button>
            <button onClick={() => sendMessage('Доступ к Jira')} style={{ padding: '6px 12px', background: '#fff', border: '1px solid #ccc', borderRadius: '20px', cursor: 'pointer', fontSize: '12px' }}>🔐 Доступ</button>
          </div>

          <div style={{ display: 'flex', gap: '10px' }}>
            <input value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && sendMessage(input)} placeholder="Напиши сообщение..." style={{ flex: 1, padding: '12px', border: '1px solid #ccc', borderRadius: '8px', fontSize: '16px' }} />
            <button onClick={() => sendMessage(input)} disabled={loading} style={{ padding: '12px 24px', background: AVITO_GREEN, border: 'none', borderRadius: '8px', fontWeight: 'bold', cursor: 'pointer' }}>Отправить</button>
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <div style={{ background: '#fff', borderRadius: '12px', padding: '20px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '15px', marginBottom: '15px' }}>
              <div style={{ width: '50px', height: '50px', borderRadius: '50%', background: AVITO_PURPLE, color: '#fff', display: 'flex', justifyContent: 'center', alignItems: 'center', fontSize: '20px', fontWeight: 'bold' }}>
                {userName.charAt(0)}
              </div>
              <div>
                <h3 style={{ margin: 0 }}>{userName}</h3>
                <p style={{ margin: 0, color: '#666', fontSize: '14px' }}>{userRole === 'intern_backend' ? 'Backend-стажёр' : userRole === 'intern_frontend' ? 'Frontend-стажёр' : 'Наставник'}</p>
              </div>
            </div>
            
            <div style={{ marginBottom: '15px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', fontSize: '14px', color: '#666' }}>
                <span>Прогресс</span>
                <span>{progress.done} / {progress.total}</span>
              </div>
              <div style={{ width: '100%', height: '8px', background: '#E0E0E0', borderRadius: '4px', overflow: 'hidden' }}>
                <div style={{ width: `${progress.percent}%`, height: '100%', background: AVITO_GREEN, transition: 'width 0.3s' }}></div>
              </div>
            </div>
          </div>

          <div style={{ background: '#fff', borderRadius: '12px', padding: '20px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)', flex: 1, overflowY: 'auto' }}>
            <h3 style={{ marginTop: 0, fontSize: '16px' }}>Мои задачи</h3>
            {tasks.length === 0 ? <p style={{ color: '#888', fontSize: '14px' }}>Все задачи выполнены!</p> : null}
            {tasks.map(task => (
              <div key={task.id} style={{ display: 'flex', gap: '10px', marginBottom: '12px', padding: '10px', background: task.is_done ? '#F0F0F0' : '#F9F9F9', borderRadius: '8px' }}>
                <input type="checkbox" checked={task.is_done} onChange={() => completeTask(task.id)} style={{ marginTop: '3px', cursor: 'pointer' }} />
                <div>
                  <strong style={{ fontSize: '13px', textDecoration: task.is_done ? 'line-through' : 'none', color: task.is_done ? '#888' : '#000' }}>{task.title}</strong>
                  <p style={{ margin: '4px 0 0', fontSize: '11px', color: '#666', textDecoration: task.is_done ? 'line-through' : 'none' }}>{task.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

export default App