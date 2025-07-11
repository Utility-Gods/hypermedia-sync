import { Hono } from 'hono'
import { cors } from 'hono/cors'
import { Effect, Ref, Stream, Queue, Fiber, Runtime, Layer, Context, Scope } from 'effect'
import { html } from 'hono/html'
import { serveStatic } from 'hono/cloudflare-workers'

// Types
interface CheckboxData {
  id: number
  checked: boolean
}

interface Connection {
  id: string
  writer: WritableStreamDefaultWriter
  done: Promise<void>
}

interface SSEEvent {
  name: string
  data: string
  excludeId?: string
}

// Services
class CheckboxService extends Context.Tag("CheckboxService")<
  CheckboxService,
  {
    readonly toggle: (id: number) => Effect.Effect<boolean>
    readonly getAll: () => Effect.Effect<readonly CheckboxData[]>
    readonly getCheckedCount: () => Effect.Effect<number>
  }
>() {}

class SSEHub extends Context.Tag("SSEHub")<
  SSEHub,
  {
    readonly register: (conn: Connection) => Effect.Effect<void>
    readonly unregister: (id: string) => Effect.Effect<void>
    readonly broadcast: (event: SSEEvent) => Effect.Effect<void>
  }
>() {}

// Implementations
const makeCheckboxService = Effect.gen(function* () {
  const checkboxes = yield* Ref.make<Map<number, boolean>>(
    new Map(Array.from({ length: 10000 }, (_, i) => [i + 1, false]))
  )

  return CheckboxService.of({
    toggle: (id: number) =>
      Ref.modify(checkboxes, (map) => {
        const newMap = new Map(map)
        const newState = !map.get(id)
        newMap.set(id, newState)
        return [newState, newMap]
      }),

    getAll: () =>
      Ref.get(checkboxes).pipe(
        Effect.map((map) =>
          Array.from(map.entries()).map(([id, checked]) => ({ id, checked }))
        )
      ),

    getCheckedCount: () =>
      Ref.get(checkboxes).pipe(
        Effect.map((map) => Array.from(map.values()).filter(Boolean).length)
      )
  })
})

const makeSSEHub = Effect.gen(function* () {
  const connections = yield* Ref.make<Map<string, Connection>>(new Map())
  const events = yield* Queue.unbounded<SSEEvent>()
  
  // Start processing events
  yield* Effect.forkDaemon(
    Stream.fromQueue(events).pipe(
      Stream.runForEach((event) =>
        Effect.gen(function* () {
          const conns = yield* Ref.get(connections)
          
          yield* Effect.forEach(
            Array.from(conns.entries()),
            ([connId, conn]) => {
              if (connId !== event.excludeId) {
                return Effect.tryPromise(async () => {
                  const encoder = new TextEncoder()
                  const eventData = event.data.replace(/\n/g, '\ndata: ')
                  const message = `event: ${event.name}\ndata: ${eventData}\n\n`
                  await conn.writer.write(encoder.encode(message))
                }).pipe(Effect.catchAll(() => Effect.void))
              }
              return Effect.void
            },
            { concurrency: "unbounded" }
          )
        })
      )
    )
  )

  return SSEHub.of({
    register: (conn: Connection) =>
      Ref.update(connections, (map) => new Map(map).set(conn.id, conn)),
      
    unregister: (id: string) =>
      Ref.update(connections, (map) => {
        const newMap = new Map(map)
        newMap.delete(id)
        return newMap
      }),
      
    broadcast: (event: SSEEvent) => Queue.offer(events, event)
  })
})

// HTML generation
const generateCheckboxHTML = (id: number, checked: boolean) => {
  const checkedAttr = checked ? 'checked' : ''
  return `<input type="checkbox" id="cb-${id}" ${checkedAttr} hx-post="/toggle/${id}" hx-swap="none"><label for="cb-${id}">${id}</label>`
}

const indexTemplate = (data: { checkboxes: CheckboxData[], checkedCount: number, originatorId: string }) => html`<!DOCTYPE html>
<html>
<head>
    <title>10,000 Checkboxes - Hypermedia Sync Experiment</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="/static/js/htmx.js"></script>
    <script src="/static/js/sse.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            /* Primary Colors - Vibrant Orange/Red */
            --color-primary-50: #fff7ed;
            --color-primary-100: #ffedd5;
            --color-primary-200: #fed7aa;
            --color-primary-300: #fdba74;
            --color-primary-400: #fb923c;
            --color-primary-500: #f97316;
            --color-primary-600: #f54a00;
            --color-primary-700: #c2410c;
            --color-primary-800: #9a3412;
            --color-primary-900: #7c2d12;
            
            /* Secondary Colors - Very Dark Navy Blue */
            --color-secondary-50: #f8fafc;
            --color-secondary-100: #f1f5f9;
            --color-secondary-200: #e2e8f0;
            --color-secondary-300: #cbd5e1;
            --color-secondary-400: #94a3b8;
            --color-secondary-500: #64748b;
            --color-secondary-600: #475569;
            --color-secondary-700: #334155;
            --color-secondary-800: #1e293b;
            --color-secondary-900: #0f172a;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, var(--color-secondary-900) 0%, var(--color-secondary-800) 100%);
            min-height: 100vh;
            color: var(--color-secondary-50);
        }
        
        .hero {
            text-align: center;
            padding: 3rem 1rem;
            color: var(--color-secondary-50);
        }
        
        .hero h1 {
            font-size: 3rem;
            font-weight: 700;
            margin-bottom: 1rem;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
            background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .hero .subtitle {
            font-size: 1.25rem;
            opacity: 0.9;
            margin-bottom: 2rem;
            font-weight: 300;
            color: var(--color-secondary-300);
        }
        
        .experiment-info {
            max-width: 800px;
            margin: 0 auto 3rem;
            padding: 2rem;
            background: rgba(30, 41, 59, 0.5);
            backdrop-filter: blur(10px);
            border-radius: 1rem;
            border: 1px solid var(--color-secondary-700);
        }
        
        .experiment-info p {
            font-size: 1.125rem;
            line-height: 1.8;
            margin-bottom: 1rem;
            color: var(--color-secondary-200);
        }
        
        .experiment-info strong {
            color: var(--color-primary-500);
        }
        
        .github-link {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.75rem 1.5rem;
            background: var(--color-primary-600);
            color: white;
            text-decoration: none;
            border-radius: 2rem;
            font-weight: 600;
            transition: transform 0.2s, box-shadow 0.2s, background 0.2s;
            box-shadow: 0 4px 6px rgba(0,0,0,0.2);
        }
        
        .github-link:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0,0,0,0.3);
            background: var(--color-primary-700);
        }
        
        .stats {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        .stats-card {
            display: inline-block;
            background: var(--color-secondary-800);
            padding: 1.5rem 3rem;
            border-radius: 1rem;
            box-shadow: 0 10px 25px rgba(0,0,0,0.3);
            border: 1px solid var(--color-secondary-700);
        }
        
        .stats-card .number {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .stats-card .label {
            font-size: 0.875rem;
            color: var(--color-secondary-400);
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        .checkbox-container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 0 1rem;
        }
        
        .checkbox-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
            gap: 0.5rem;
            max-height: 60vh;
            overflow-y: auto;
            padding: 1.5rem;
            background: var(--color-secondary-800);
            border-radius: 1rem;
            box-shadow: 0 10px 25px rgba(0,0,0,0.3);
            border: 1px solid var(--color-secondary-700);
        }
        
        .checkbox-grid::-webkit-scrollbar {
            width: 12px;
        }
        
        .checkbox-grid::-webkit-scrollbar-track {
            background: var(--color-secondary-700);
            border-radius: 10px;
        }
        
        .checkbox-grid::-webkit-scrollbar-thumb {
            background: var(--color-primary-600);
            border-radius: 10px;
        }
        
        .checkbox-grid::-webkit-scrollbar-thumb:hover {
            background: var(--color-primary-700);
        }
        
        .checkbox-item {
            display: flex;
            align-items: center;
            padding: 0.5rem;
            background: var(--color-secondary-700);
            border-radius: 0.5rem;
            transition: all 0.2s;
            border: 1px solid var(--color-secondary-600);
        }
        
        .checkbox-item:hover {
            background: var(--color-secondary-600);
            transform: scale(1.05);
            border-color: var(--color-primary-600);
        }
        
        .checkbox-item input[type="checkbox"] {
            width: 18px;
            height: 18px;
            margin-right: 0.5rem;
            cursor: pointer;
            accent-color: var(--color-primary-600);
        }
        
        .checkbox-item label {
            cursor: pointer;
            font-size: 0.875rem;
            color: var(--color-secondary-200);
            user-select: none;
        }
        
        .footer {
            text-align: center;
            padding: 3rem 1rem;
            color: var(--color-secondary-400);
        }
        
        @media (max-width: 768px) {
            .hero h1 {
                font-size: 2rem;
            }
            
            .checkbox-grid {
                grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
                max-height: 50vh;
            }
        }
    </style>
</head>
<body>
    <div class="hero">
        <h1>10,000 Checkboxes</h1>
        <p class="subtitle">A Real-Time Hypermedia Experiment</p>
        
        <div class="experiment-info">
            <p>
                This is an experiment in <strong>hypermedia-driven real-time synchronization</strong>. 
                Every checkbox you click is instantly synchronized across all connected browsers using 
                Server-Sent Events (SSE) and HTMX.
            </p>
            <p>
                Unlike traditional approaches that send JSON and require client-side rendering, 
                this demo sends pure HTML fragments. Each checkbox update is a tiny ~50 byte HTML snippet 
                that surgically updates just that checkbox across all browsers.
            </p>
            <p>
                Open this page in multiple tabs or share with friends to see the magic! 
                The server maintains zero client state - it's all just HTML over the wire.
            </p>
            <a href="https://github.com/Utility-Gods/hypermedia-sync" class="github-link" target="_blank">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>
    
    <div class="stats">
        <div class="stats-card">
            <div class="number" id="checked-count">${data.checkedCount}</div>
            <div class="label">Checkboxes Checked</div>
        </div>
    </div>

    <div class="checkbox-container">
        <!-- SSE Connection Wrapper -->
        <div hx-ext="sse" 
             sse-connect="/events?originator=${data.originatorId}" 
             id="sse-wrapper">
            <div class="checkbox-grid" id="team-section">
                ${data.checkboxes.map(cb => `
                <div class="checkbox-item" id="checkbox-${cb.id}" 
                     sse-swap="checkbox-${cb.id}-updated"
                     hx-swap="innerHTML">
                    <input type="checkbox" 
                           id="cb-${cb.id}" 
                           ${cb.checked ? 'checked' : ''}
                           hx-post="/toggle/${cb.id}"
                           hx-swap="none">
                    <label for="cb-${cb.id}">${cb.id}</label>
                </div>
                `).join('')}
            </div>
        </div>
    </div>
    
    <div class="footer">
        <p>Built with HTMX + Server-Sent Events • No WebSockets, No JSON, Just HTML</p>
    </div>

    <script>
        // Server-generated originator ID
        window.originatorId = '${data.originatorId}';
        
        // Add originator ID to all HTMX requests
        document.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.headers['X-Originator-ID'] = window.originatorId;
        });

        // Update checked count on page updates
        document.addEventListener('htmx:afterSwap', function(evt) {
            updateCheckedCount();
        });

        function updateCheckedCount() {
            const checked = document.querySelectorAll('input[type="checkbox"]:checked').length;
            const countElement = document.getElementById('checked-count');
            if (countElement) {
                countElement.textContent = checked;
            }
        }

        // Initial count
        updateCheckedCount();
    </script>
</body>
</html>`

// Main application
const app = new Hono()

// Create services layer
const MainLayer = Layer.scopedDiscard(
  Effect.gen(function* () {
    const checkboxService = yield* makeCheckboxService
    const sseHub = yield* makeSSEHub
    
    return Layer.mergeAll(
      Layer.succeed(CheckboxService, checkboxService),
      Layer.succeed(SSEHub, sseHub)
    )
  })
).pipe(Layer.provide(Scope.Scope))

// Runtime
const runtime = Runtime.defaultRuntime.pipe(
  Runtime.withLayer(MainLayer)
)

// Middleware
app.use('*', cors())
app.use('/static/*', serveStatic({ root: './' }))

// Routes
app.get('/', async (c) => {
  const program = Effect.gen(function* () {
    const checkboxService = yield* CheckboxService
    const checkboxes = yield* checkboxService.getAll()
    const checkedCount = yield* checkboxService.getCheckedCount()
    
    const originatorId = `page-${Date.now()}-${Math.floor(Math.random() * 1000000)}`
    
    return c.html(indexTemplate({
      checkboxes: Array.from(checkboxes),
      checkedCount,
      originatorId
    }))
  })
  
  return Runtime.runPromise(runtime)(program)
})

app.get('/events', async (c) => {
  const originatorId = c.req.query('originator') || `sse-${Date.now()}`
  
  const { readable, writable } = new TransformStream()
  const writer = writable.getWriter()
  
  const encoder = new TextEncoder()
  
  // Set SSE headers
  c.header('Content-Type', 'text/event-stream')
  c.header('Cache-Control', 'no-cache')
  c.header('Connection', 'keep-alive')
  
  const program = Effect.gen(function* () {
    const sseHub = yield* SSEHub
    
    const conn: Connection = {
      id: originatorId,
      writer,
      done: new Promise(() => {})
    }
    
    yield* sseHub.register(conn)
    
    // Keep connection alive with ping
    const keepAlive = yield* Effect.forkDaemon(
      Effect.repeat(
        Effect.tryPromise(() => 
          writer.write(encoder.encode(':ping\n\n'))
        ).pipe(Effect.catchAll(() => Effect.void)),
        { schedule: "30 seconds" }
      )
    )
    
    // Wait for client disconnect
    yield* Effect.async<void>((resume) => {
      c.req.raw.signal.addEventListener('abort', () => {
        resume(Effect.void)
      })
    })
    
    yield* Fiber.interrupt(keepAlive)
    yield* sseHub.unregister(originatorId)
    yield* Effect.tryPromise(() => writer.close())
  })
  
  Runtime.runPromise(runtime)(program).catch(console.error)
  
  return c.body(readable)
})

app.post('/toggle/:id', async (c) => {
  const id = parseInt(c.req.param('id'))
  const originatorId = c.req.header('X-Originator-ID')
  
  if (id < 1 || id > 10000) {
    return c.text('Invalid checkbox ID', 400)
  }
  
  const program = Effect.gen(function* () {
    const checkboxService = yield* CheckboxService
    const sseHub = yield* SSEHub
    
    const newState = yield* checkboxService.toggle(id)
    const checkboxHTML = generateCheckboxHTML(id, newState)
    
    yield* sseHub.broadcast({
      name: `checkbox-${id}-updated`,
      data: checkboxHTML,
      excludeId: originatorId
    })
    
    return c.body(null, 204)
  })
  
  return Runtime.runPromise(runtime)(program)
})

export default app