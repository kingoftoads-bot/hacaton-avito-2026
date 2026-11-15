import React, { useState, useEffect } from 'react'
import Markdown from 'react-markdown'

const AVITO_GREEN = '#97EB33'
const AVITO_BLUE = '#009CF0'
const AVITO_PURPLE = '#965EEB'

type ChatMessage = {
  role: 'user' | 'bot'
  text: string
}

type Task = {
  id: number
  title: string
  description: string
  is_done: boolean
}

function App() {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'bot', text: 'Привет! Я твой ИИ-наставник. Выбери свою роль вверху, чтобы я знал, как тебе помочь.' }
  ])
  const [loading, setLoading] = useState(false)
  const [userRole, setUserRole] = useState('intern_backend')
  const [progress, setProgress] = useState({ total: 0, done: 0, percent: 0 })
  const [tasks, setTasks] = useState<Task[]>([])

  // Загружаем прогресс и задачи при смене роли или после обновления
  const loadData = () => {
    fetch(`http://localhost:8080/api/progress?role=${userRole}`)
      .then(res => res.json())
      .then(data => setProgress(data))

    // Загружаем задачи (в реальном проекте это отдельный эндпоинт, 
    // но для простоты мы просто достанем их из ответа чата или сделаем запрос)
    fetch(`http://localhost:8080/api/tasks?role=${userRole}`)
      .then(res => res.json())
      .then(data => setTasks(data || []))
      .catch(() => {})
  }

  useEffect(() => {
    loadData()
  }, [userRole])

  const completeTask = async (id: number) => {
    await fetch(`http://localhost:8080/api/tasks/${id}/complete`, { method: 'POST' })
    loadData() // Обновляем прогресс и список
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
        body: JSON.stringify({ message: text, role: userRole })
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
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <div style={{ display: 'flex' }}>
              <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_GREEN }}></div>
              <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_BLUE, marginLeft: '-10px' }}></div>
              <div style={{ width: '25px', height: '25px', borderRadius: '50%', background: AVITO_PURPLE, marginLeft: '-10px' }}></div>
            </div>
            <h1 style={{ margin: 0, fontSize: '24px' }}>AI-buddy</h1>
          </div>
          
          <select 
            value={userRole} 
            onChange={(e) => setUserRole(e.target.value)}
            style={{ padding: '8px 12px', borderRadius: '8px', border: '1px solid #ccc', fontSize: '14px', cursor: 'pointer' }}
          >
            <option value="intern_backend">Стажёр Backend</option>
            <option value="intern_frontend">Стажёр Frontend</option>
            <option value="mentor">Наставник</option>
          </select>
        </div>

        {/* Полоса прогресса */}
        <div style={{ marginBottom: '20px', background: '#fff', padding: '15px', borderRadius: '12px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', fontSize: '14px', color: '#666' }}>
            <span>Прогресс адаптации</span>
            <span>{progress.done} из {progress.total} задач</span>
          </div>
          <div style={{ width: '100%', height: '10px', background: '#E0E0E0', borderRadius: '5px', overflow: 'hidden' }}>
            <div style={{ width: `${progress.percent}%`, height: '100%', background: AVITO_GREEN, transition: 'width 0.3s' }}></div>
          </div>
        </div>

        {/* Блок задач (трекер) */}
        {tasks.length > 0 && (
          <div style={{ marginBottom: '20px', background: '#fff', padding: '15px', borderRadius: '12px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
            <h3 style={{ marginTop: 0, fontSize: '16px' }}>Мои задачи на неделю</h3>
            {tasks.map(task => (
              <div key={task.id} style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '10px', padding: '10px', background: '#F9F9F9', borderRadius: '8px' }}>
                <input 
                  type="checkbox" 
                  checked={task.is_done} 
                  onChange={() => completeTask(task.id)}
                  style={{ cursor: 'pointer', width: '18px', height: '18px' }}
                />
                <div style={{ flex: 1 }}>
                  <strong style={{ fontSize: '14px', textDecoration: task.is_done ? 'line-through' : 'none' }}>{task.title}</strong>
                  <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>{task.description}</p>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Область чата */}
        <div style={{ background: '#fff', borderRadius: '12px', padding: '20px', minHeight: '300px', maxHeight: '400px', overflowY: 'auto', marginBottom: '20px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}>
          {messages.map((msg, i) => (
            <div key={i} style={{ marginBottom: '15px', display: 'flex', justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start' }}>
              <div style={{ 
                padding: '12px 16px', borderRadius: '12px', maxWidth: '80%',
                background: msg.role === 'user' ? AVITO_GREEN : '#F0F0F0',
                color: '#000'
              }}>
                {msg.role === 'bot' ? <Markdown>{msg.text}</Markdown> : msg.text}
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