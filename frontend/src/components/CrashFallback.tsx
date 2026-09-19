export function CrashFallback() {
  return (
    <main role="alert" style={{ padding: '24px', fontFamily: 'sans-serif' }}>
      <h1>Migration Lab encountered an error</h1>
      <p>Reload the page to try again.</p>
      <button onClick={() => window.location.reload()}>Reload</button>
    </main>
  )
}
