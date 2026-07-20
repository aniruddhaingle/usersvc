import { useEffect, useState } from 'react'

const api = {
  async listUsers() {
    const res = await fetch('/api/users')
    if (res.ok) return res.json()
    throw new Error(`List failed (HTTP ${res.status})`)
  },
  async createUser(email, password) {
    const res = await fetch('/api/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (res.status === 201) return res.json()
    if (res.status === 409) throw new Error('That email is already registered.')
    throw new Error(`Create failed (HTTP ${res.status})`)
  },
  async getUser(id) {
    const res = await fetch(`/api/users/${encodeURIComponent(id)}`)
    if (res.ok) return res.json()
    if (res.status === 404) throw new Error('No user with that id.')
    throw new Error(`Lookup failed (HTTP ${res.status})`)
  },
  async deleteUser(id) {
    const res = await fetch(`/api/users/${encodeURIComponent(id)}`, { method: 'DELETE' })
    if (res.status === 204 || res.status === 404) return
    throw new Error(`Delete failed (HTTP ${res.status})`)
  },
}

function Status({ status }) {
  if (!status) return null
  return <p className={`status ${status.kind}`}>{status.text}</p>
}

function fmtDate(iso) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

export default function App() {
  const [users, setUsers] = useState([])
  const [listStatus, setListStatus] = useState(null)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [createStatus, setCreateStatus] = useState(null)
  const [creating, setCreating] = useState(false)

  const [lookupId, setLookupId] = useState('')
  const [lookupStatus, setLookupStatus] = useState(null)
  const [looking, setLooking] = useState(false)

  async function refresh() {
    try {
      setUsers(await api.listUsers())
      setListStatus(null)
    } catch (err) {
      setListStatus({ kind: 'err', text: err.message })
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  async function onCreate(e) {
    e.preventDefault()
    setCreating(true)
    setCreateStatus(null)
    try {
      const user = await api.createUser(email.trim(), password)
      setCreateStatus({ kind: 'ok', text: `Created ${user.email}` })
      setEmail('')
      setPassword('')
      refresh()
    } catch (err) {
      setCreateStatus({ kind: 'err', text: err.message })
    } finally {
      setCreating(false)
    }
  }

  async function onLookup(e) {
    e.preventDefault()
    setLooking(true)
    setLookupStatus(null)
    try {
      const user = await api.getUser(lookupId.trim())
      setLookupStatus({
        kind: 'ok',
        text: `Found ${user.email} (created ${fmtDate(user.created_at)})`,
      })
      setLookupId('')
    } catch (err) {
      setLookupStatus({ kind: 'err', text: err.message })
    } finally {
      setLooking(false)
    }
  }

  async function onDelete(id) {
    try {
      await api.deleteUser(id)
    } catch (err) {
      alert(err.message)
    }
    refresh()
  }

  return (
    <div className="page">
      <header>
        <h1>
          user<span className="accent">svc</span>
        </h1>
        <p className="tagline">users API &middot; create, look up, delete</p>
      </header>

      <main>
        <div className="cards">
          <section className="card">
            <h2>Create user</h2>
            <form onSubmit={onCreate}>
              <label>
                Email
                <input
                  type="email"
                  required
                  placeholder="ada@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </label>
              <label>
                Password
                <input
                  type="password"
                  required
                  minLength={6}
                  placeholder="at least 6 characters"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </label>
              <button type="submit" disabled={creating}>
                {creating ? 'Creating…' : 'Create'}
              </button>
            </form>
            <Status status={createStatus} />
          </section>

          <section className="card">
            <h2>Look up by id</h2>
            <form onSubmit={onLookup}>
              <label>
                User id
                <input
                  required
                  placeholder="uuid, e.g. 38198ec3-…"
                  value={lookupId}
                  onChange={(e) => setLookupId(e.target.value)}
                />
              </label>
              <button type="submit" disabled={looking}>
                {looking ? 'Looking…' : 'Look up'}
              </button>
            </form>
            <Status status={lookupStatus} />
          </section>
        </div>

        <section className="card wide">
          <h2>All users</h2>
          <Status status={listStatus} />
          {users.length === 0 ? (
            <p className="empty">No users yet — create one above.</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Email</th>
                  <th>Id</th>
                  <th>Created</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id}>
                    <td>{u.email}</td>
                    <td>
                      <code
                        title="click to copy"
                        onClick={() => navigator.clipboard?.writeText(u.id)}
                      >
                        {u.id}
                      </code>
                    </td>
                    <td>{fmtDate(u.created_at)}</td>
                    <td>
                      <button className="danger" onClick={() => onDelete(u.id)}>
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </section>
      </main>

      <footer>
        API: <code>GET /users</code> &middot; <code>POST /users</code> &middot;{' '}
        <code>GET /users/&#123;id&#125;</code> &middot;{' '}
        <code>DELETE /users/&#123;id&#125;</code> &middot; <code>GET /health</code>
      </footer>
    </div>
  )
}
