import type {RetryVerdict, RuntimeResult, RuntimeVerdict} from '../api/migrationApi'
import {useMigrationStore} from '../store/migrationStore'
import './runtimePopup.css'

const VERDICT_COLOR: Record<RuntimeVerdict, string> = {
  COMPLETED: '#4caf82',
  EXCEEDS_DEADLINE: '#e8a33d',
  FAILED: '#e5646e',
}

const RETRY_COLOR: Record<RetryVerdict, string> = {
  SAFE_TO_RETRY: '#4caf82',
  NEEDS_MANUAL_CLEANUP: '#e8a33d',
  FAILURE_LOOP: '#e5646e',
  NOT_APPLICABLE: '#98c1d9',
}

function Badge({label, color}: {label: string; color: string}) {
  return <span className="runtime-badge" style={{color}}>{label}</span>
}

function Result({result}: {result: RuntimeResult}) {
  return (
    <>
      <div className="runtime-verdicts">
        <Badge label={result.verdict} color={VERDICT_COLOR[result.verdict]} />
        <Badge label={`retry: ${result.retry}`} color={RETRY_COLOR[result.retry]} />
        <Badge label={`deadline ${result.deadlineMs} ms`} color="#98c1d9" />
      </div>

      <p className="runtime-message">{result.message}</p>

      {result.findings.length > 0 && (
        <section className="runtime-section">
          <h3>Findings</h3>
          <ul>
            {result.findings.map((finding, index) => (
              <li key={index} className="runtime-finding">
                <strong>{finding.type}</strong> {finding.message}
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className="runtime-section">
        <h3>Seeded</h3>
        <ul>
          {result.seeded.map((table) => (
            <li key={table.table}>
              {table.table} — {table.error ? table.error : `${table.rows.toLocaleString()} rows`}
            </li>
          ))}
          {result.seeded.length === 0 && <li>no table needed seeding</li>}
        </ul>
      </section>

      <section className="runtime-section">
        <h3>Statements</h3>
        <ul>
          {result.statements.map((statement) => (
            <li key={statement.statementIndex}>
              {statement.durationMs} ms · {statement.verdict}
              {statement.strongestLock && ` · ${statement.strongestLock}`}
              <code className="runtime-sql">{statement.sql}</code>
            </li>
          ))}
        </ul>
      </section>

      {result.probes.length > 0 && (
        <section className="runtime-section">
          <h3>Concurrent readers</h3>
          <ul>
            {result.probes.map((probe) => (
              <li key={probe.table}>
                {probe.table} — blocked {probe.blockedMs} ms, slowest read {probe.maxLatencyMs} ms,
                {' '}{probe.samples} reads
              </li>
            ))}
          </ul>
        </section>
      )}
    </>
  )
}

/** Result of a runtime analysis, shown over the timeline while it runs and once it is done. */
export function RuntimePopup() {
  const target = useMigrationStore((state) => state.runtimeTarget)
  const result = useMigrationStore((state) => state.runtimeResult)
  const error = useMigrationStore((state) => state.runtimeError)
  const closeRuntime = useMigrationStore((state) => state.closeRuntime)

  if (!target && !result && !error) return null

  return (
    <div className="runtime-backdrop" onClick={target ? undefined : closeRuntime}>
      <article className="runtime-panel" onClick={(event) => event.stopPropagation()}>
        <div className="runtime-head">
          <h2>runtime · {result?.migrationId ?? target}</h2>
          {!target && <button className="runtime-close" onClick={closeRuntime}>close</button>}
        </div>

        {target && <p className="runtime-message">Seeding a container and measuring {target}…</p>}
        {error && <p className="runtime-error">{error}</p>}
        {result && <Result result={result} />}
      </article>
    </div>
  )
}
