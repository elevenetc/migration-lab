import {Fragment, useState} from 'react'
import type {RetryVerdict, RuntimeAnalysisResult, RuntimeMigrationResult, RuntimeVerdict, StatementMeasurement} from '../api/migrationApi'
import {useMigrationStore} from '../store/migrationStore'
import {summarizeRuntime} from './summarizeRuntime'
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

/**
 * What the statement cost: what static analysis predicted, and what the run measured. The two
 * disagreeing is the point — the prediction is a guess the measurement scores.
 */
function Classes({statement}: {statement: StatementMeasurement}) {
  const {predictedClass, observedClass} = statement
  if (!predictedClass && !observedClass) return null

  const mismatch = Boolean(predictedClass && observedClass && predictedClass !== observedClass)
  return (
    <span className={mismatch ? 'runtime-classes runtime-classes-mismatch' : 'runtime-classes'}>
      {predictedClass && ` · predicted ${predictedClass}`}
      {observedClass && ` · observed ${observedClass}`}
    </span>
  )
}

function Result({result}: {result: RuntimeMigrationResult}) {
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
              <Classes statement={statement} />
              {statement.rewrittenRelations.length > 0 && ` · rewrote ${statement.rewrittenRelations.join(', ')}`}
              {statement.tuplesRead > 0 && ` · read ${statement.tuplesRead.toLocaleString()} rows`}
              <code className="runtime-sql">{statement.sql}</code>
            </li>
          ))}
        </ul>
      </section>
    </>
  )
}

/** Every measured migration, one row each; a row expands into its full result. */
function MigrationTable({migrations}: {migrations: RuntimeMigrationResult[]}) {
  const [expanded, setExpanded] = useState<string | null>(null)

  return (
    <table className="runtime-timeline">
      <thead>
        <tr>
          <th>migration</th>
          <th>verdict</th>
          <th>ms</th>
          <th>lock</th>
          <th>class</th>
        </tr>
      </thead>
      <tbody>
        {migrations.map((migration) => {
          const summary = summarizeRuntime(migration)
          const open = expanded === migration.migrationId
          return (
            <Fragment key={migration.migrationId}>
              <tr
                className={open ? 'runtime-timeline-row runtime-timeline-row-open' : 'runtime-timeline-row'}
                onClick={() => setExpanded(open ? null : migration.migrationId)}
              >
                <td>{migration.version}</td>
                <td><Badge label={migration.verdict} color={VERDICT_COLOR[migration.verdict]} /></td>
                <td>{summary.durationMs}</td>
                <td>{summary.strongestLock ?? '—'}</td>
                <td>{summary.worstClass ?? '—'}</td>
              </tr>
              {open && (
                <tr>
                  <td colSpan={5} className="runtime-timeline-detail">
                    <Result result={migration} />
                  </td>
                </tr>
              )}
            </Fragment>
          )
        })}
      </tbody>
    </table>
  )
}

/** One measured migration reads as its full result; several as a table, so a regression stands out. */
function Analysis({analysis}: {analysis: RuntimeAnalysisResult}) {
  if (analysis.migrations.length === 0) {
    return <p className="runtime-error">{analysis.message}</p>
  }
  if (analysis.migrations.length === 1 && !analysis.stopped) {
    return <Result result={analysis.migrations[0]} />
  }
  return (
    <>
      <div className="runtime-verdicts">
        <Badge label={`deadline ${analysis.deadlineMs} ms`} color="#98c1d9" />
        <Badge label={`${analysis.rows.toLocaleString()} rows`} color="#98c1d9" />
      </div>
      <p className={analysis.stopped ? 'runtime-error' : 'runtime-message'}>{analysis.message}</p>
      <MigrationTable migrations={analysis.migrations} />
    </>
  )
}

/** Result of a runtime analysis, shown over the timeline while it runs and once it is done. */
export function RuntimePopup() {
  const running = useMigrationStore((state) => state.runtimeRunning)
  const target = useMigrationStore((state) => state.runtimeTarget)
  const analysis = useMigrationStore((state) => state.runtimeAnalysis)
  const error = useMigrationStore((state) => state.runtimeError)
  const migrationCount = useMigrationStore((state) => state.migrations.length)
  const closeRuntime = useMigrationStore((state) => state.closeRuntime)

  if (!running && !analysis && !error) return null

  return (
    <div className="runtime-backdrop" onClick={running ? undefined : closeRuntime}>
      <article className="runtime-panel" onClick={(event) => event.stopPropagation()}>
        <div className="runtime-head">
          <h2>runtime · {target ?? 'all migrations'}</h2>
          {!running && <button className="runtime-close" onClick={closeRuntime}>close</button>}
        </div>

        {running && <p className="runtime-message">Seeding a container and measuring {target ?? `${migrationCount} migrations`}…</p>}
        {error && <p className="runtime-error">{error}</p>}
        {analysis && <Analysis analysis={analysis} />}
      </article>
    </div>
  )
}
