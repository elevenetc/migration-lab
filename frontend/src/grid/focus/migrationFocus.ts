import type {FocusCell, FocusLayout} from 'active-grid';
import {MIGRATION_KIND_COLOR, type MigrationData, migrationDrawer} from '../drawers/migration';

function migrationOf(layers: FocusCell['layers']): MigrationData | null {
    const layer = layers.find(({drawer}) => drawer === migrationDrawer);
    return layer ? (layer.data as MigrationData) : null;
}

/** The whole migration behind a cell: the SQL the backend parsed, scrollable and copyable. */
export const migrationFocus: FocusLayout = ({layers}) => {
    const migration = migrationOf(layers);
    if (!migration) return null;

    const panel = document.createElement('article');
    panel.className = 'migration-focus';
    panel.style.setProperty('--kind-color', MIGRATION_KIND_COLOR[migration.kind]);

    const title = document.createElement('h2');
    title.textContent = migration.version;
    panel.appendChild(title);

    const subtitle = document.createElement('p');
    subtitle.className = 'migration-focus-table';
    const count = migration.sql.length;
    subtitle.textContent = `${migration.kind} · ${migration.table} · ${count} ${count === 1 ? 'statement' : 'statements'}`;
    panel.appendChild(subtitle);

    migration.warnings.forEach(({title: heading, message}) => {
        const warning = document.createElement('aside');
        warning.className = 'migration-focus-lock';
        const name = document.createElement('strong');
        name.textContent = heading;
        warning.appendChild(name);
        const detail = document.createElement('p');
        detail.textContent = message;
        warning.appendChild(detail);
        panel.appendChild(warning);
    });

    const body = document.createElement('pre');
    body.textContent = migration.sql.join('\n\n');
    panel.appendChild(body);

    return panel;
};
