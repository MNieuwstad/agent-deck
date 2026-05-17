// CreateSessionDialog.js -- Modal form for creating a new session.
//
// PR-1 changes (feature-parity with TUI, phase 1):
//   • TOOLS expanded from 4 to 7: adds opencode, copilot, pi.
//   • GROUP selector added — reads existing groups from menuModelSignal and
//     sends groupPath in the POST body when a group is chosen.
//   • seg-row gains flex-wrap so 7 buttons fit within the 560 px dialog.
//
// Restyled (PR-B) to use the bundle's .dialog / .dh / .db / .df /
// .field / .seg-row / .btn classes from app.css.
import { html } from 'htm/preact'
import { useState } from 'preact/hooks'
import { createSessionDialogSignal, mutationsEnabledSignal } from './state.js'
import { menuModelSignal } from './dataModel.js'
import { Icon, ICONS } from './icons.js'
import { apiFetch } from './api.js'

// All seven built-in tools recognised by the backend (session.Instance.Tool).
// Custom tools configured in userconfig.toml are not enumerated here; they
// can still be typed into a future free-text field. The seg-row wraps so
// all buttons stay accessible on the 560 px max-width dialog.
const TOOLS = ['claude', 'codex', 'gemini', 'opencode', 'copilot', 'pi', 'shell']

export function CreateSessionDialog() {
  const [title, setTitle] = useState('')
  const [tool, setTool] = useState('claude')
  const [path, setPath] = useState('')
  const [group, setGroup] = useState('')
  const [error, setError] = useState(null)
  const [submitting, setSubmitting] = useState(false)

  // WEB-P0-4 prevention layer: when mutations are disabled (server
  // webMutations=false), do not render the dialog at all. Hooks order is
  // preserved by placing this guard AFTER all useState calls.
  if (!mutationsEnabledSignal.value) return null

  // Groups derived from the live SSE snapshot.  Empty array until the
  // first /api/menu response lands; the select shows only "No group" then.
  const { groups } = menuModelSignal.value

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const body = { title, tool, projectPath: path }
      // Only include groupPath when a group is explicitly selected so the
      // server-side CreateSession path receives an empty string for ungrouped
      // sessions, matching the TUI's "no group" behaviour.
      if (group) body.groupPath = group
      await apiFetch('POST', '/api/sessions', body)
      createSessionDialogSignal.value = false
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  const close = () => (createSessionDialogSignal.value = false)
  const handleBackdropClick = (e) => { if (e.target === e.currentTarget) close() }

  return html`
    <div class="overlay" onClick=${handleBackdropClick}>
      <form class="dialog" onClick=${e => e.stopPropagation()} onSubmit=${handleSubmit}>
        <div class="dh">
          <span class="kicker">NEW</span>
          <div class="t">New session</div>
          <button type="button" class="icon-btn" onClick=${close} aria-label="Close">
            <${Icon} d=${ICONS.x}/>
          </button>
        </div>
        <div class="db">
          <div class="field">
            <label>GROUP</label>
            <select value=${group} onChange=${e => setGroup(e.target.value)}>
              <option value="">— No group —</option>
              ${groups.map(g => html`
                <option key=${g.path} value=${g.path}>${g.label}</option>
              `)}
            </select>
          </div>
          <div class="field">
            <label>TITLE</label>
            <input autofocus required value=${title} onInput=${e => setTitle(e.target.value)} placeholder="my-session"/>
          </div>
          <div class="field">
            <label>WORKING DIR</label>
            <input required value=${path} onInput=${e => setPath(e.target.value)} placeholder="/absolute/path/to/project"/>
          </div>
          <div class="field">
            <label>TOOL</label>
            <div class="seg-row" style="flex-wrap: wrap;">
              ${TOOLS.map(t => html`
                <button type="button" key=${t}
                        class=${`seg-btn ${tool === t ? 'on' : ''}`}
                        onClick=${() => setTool(t)}>${t}</button>
              `)}
            </div>
          </div>
          ${error && html`
            <div style="font-family: var(--mono); font-size: 11.5px; color: var(--tn-red); padding: 8px 10px;
                        border: 1px solid rgba(247,118,142,0.3); border-radius: 4px; background: rgba(247,118,142,0.06);">
              ${error}
            </div>
          `}
        </div>
        <div class="df">
          <button type="button" class="btn ghost" onClick=${close}>Cancel</button>
          <button type="submit" class="btn primary" disabled=${submitting || !title || !path}>
            ${submitting ? 'Creating…' : html`Create session <span class="kbd">⏎</span>`}
          </button>
        </div>
      </form>
    </div>
  `
}
