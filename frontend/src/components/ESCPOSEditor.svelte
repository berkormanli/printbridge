<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { EditorState } from '@codemirror/state';
    import { EditorView, keymap, lineNumbers, highlightActiveLineGutter, highlightSpecialChars, drawSelection, highlightActiveLine } from '@codemirror/view';
    import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
    import { oneDark } from '@codemirror/theme-one-dark';

    let { value = $bindable(''), placeholder = '', onchange = () => {} } = $props();
    
    let editorContainer: HTMLDivElement;
    let editorView: EditorView | null = null;
    let commandLog = $state<{name: string, bytes: string, hex: string}[]>([]);
    let selectedCategory = $state('basic');

    // Comprehensive ESC/POS command library organized by category
    const commandCategories = {
        basic: {
            label: 'Basic',
            commands: [
                { label: 'Initialize', code: '\\x1B\\x40', desc: 'ESC @ - Reset printer', bytes: [0x1B, 0x40] },
                { label: 'Newline', code: '\\n', desc: 'LF - Line feed', bytes: [0x0A] },
                { label: 'Carriage Return', code: '\\r', desc: 'CR', bytes: [0x0D] },
                { label: 'Tab', code: '\\t', desc: 'HT - Horizontal tab', bytes: [0x09] },
            ]
        },
        text: {
            label: 'Text Style',
            commands: [
                { label: 'Bold ON', code: '\\x1B\\x45\\x01', desc: 'ESC E 1', bytes: [0x1B, 0x45, 0x01] },
                { label: 'Bold OFF', code: '\\x1B\\x45\\x00', desc: 'ESC E 0', bytes: [0x1B, 0x45, 0x00] },
                { label: 'Underline 1-dot', code: '\\x1B\\x2D\\x01', desc: 'ESC - 1', bytes: [0x1B, 0x2D, 0x01] },
                { label: 'Underline 2-dot', code: '\\x1B\\x2D\\x02', desc: 'ESC - 2', bytes: [0x1B, 0x2D, 0x02] },
                { label: 'Underline OFF', code: '\\x1B\\x2D\\x00', desc: 'ESC - 0', bytes: [0x1B, 0x2D, 0x00] },
                { label: 'Italic ON', code: '\\x1B\\x34', desc: 'ESC 4', bytes: [0x1B, 0x34] },
                { label: 'Italic OFF', code: '\\x1B\\x35', desc: 'ESC 5', bytes: [0x1B, 0x35] },
                { label: 'Reverse ON', code: '\\x1D\\x42\\x01', desc: 'GS B 1', bytes: [0x1D, 0x42, 0x01] },
                { label: 'Reverse OFF', code: '\\x1D\\x42\\x00', desc: 'GS B 0', bytes: [0x1D, 0x42, 0x00] },
            ]
        },
        align: {
            label: 'Alignment',
            commands: [
                { label: 'Left', code: '\\x1B\\x61\\x00', desc: 'ESC a 0', bytes: [0x1B, 0x61, 0x00] },
                { label: 'Center', code: '\\x1B\\x61\\x01', desc: 'ESC a 1', bytes: [0x1B, 0x61, 0x01] },
                { label: 'Right', code: '\\x1B\\x61\\x02', desc: 'ESC a 2', bytes: [0x1B, 0x61, 0x02] },
            ]
        },
        size: {
            label: 'Size',
            commands: [
                { label: 'Normal', code: '\\x1B\\x21\\x00', desc: 'ESC ! 0', bytes: [0x1B, 0x21, 0x00] },
                { label: '2x Width', code: '\\x1B\\x21\\x20', desc: 'ESC ! 32', bytes: [0x1B, 0x21, 0x20] },
                { label: '2x Height', code: '\\x1B\\x21\\x10', desc: 'ESC ! 16', bytes: [0x1B, 0x21, 0x10] },
                { label: '2x Both', code: '\\x1B\\x21\\x30', desc: 'ESC ! 48', bytes: [0x1B, 0x21, 0x30] },
                { label: '3x Both', code: '\\x1D\\x21\\x22', desc: 'GS ! 34', bytes: [0x1D, 0x21, 0x22] },
                { label: '4x Both', code: '\\x1D\\x21\\x33', desc: 'GS ! 51', bytes: [0x1D, 0x21, 0x33] },
            ]
        },
        font: {
            label: 'Font',
            commands: [
                { label: 'Font A', code: '\\x1B\\x4D\\x00', desc: 'ESC M 0 - 12x24', bytes: [0x1B, 0x4D, 0x00] },
                { label: 'Font B', code: '\\x1B\\x4D\\x01', desc: 'ESC M 1 - 9x17', bytes: [0x1B, 0x4D, 0x01] },
                { label: 'Font C', code: '\\x1B\\x4D\\x02', desc: 'ESC M 2 - 9x24', bytes: [0x1B, 0x4D, 0x02] },
            ]
        },
        feed: {
            label: 'Feed',
            commands: [
                { label: 'Feed 1 line', code: '\\x1B\\x64\\x01', desc: 'ESC d 1', bytes: [0x1B, 0x64, 0x01] },
                { label: 'Feed 2 lines', code: '\\x1B\\x64\\x02', desc: 'ESC d 2', bytes: [0x1B, 0x64, 0x02] },
                { label: 'Feed 3 lines', code: '\\x1B\\x64\\x03', desc: 'ESC d 3', bytes: [0x1B, 0x64, 0x03] },
                { label: 'Feed 5 lines', code: '\\x1B\\x64\\x05', desc: 'ESC d 5', bytes: [0x1B, 0x64, 0x05] },
                { label: 'Reverse Feed', code: '\\x1B\\x4B\\x02', desc: 'ESC K 2', bytes: [0x1B, 0x4B, 0x02] },
            ]
        },
        cut: {
            label: 'Cut',
            commands: [
                { label: 'Full Cut', code: '\\x1D\\x56\\x00', desc: 'GS V 0', bytes: [0x1D, 0x56, 0x00] },
                { label: 'Partial Cut', code: '\\x1D\\x56\\x01', desc: 'GS V 1', bytes: [0x1D, 0x56, 0x01] },
                { label: 'Feed & Cut', code: '\\x1D\\x56\\x42\\x03', desc: 'GS V 66 3', bytes: [0x1D, 0x56, 0x42, 0x03] },
            ]
        },
        cash: {
            label: 'Cash Drawer',
            commands: [
                { label: 'Open Drawer 1', code: '\\x1B\\x70\\x00\\x19\\xFA', desc: 'ESC p 0', bytes: [0x1B, 0x70, 0x00, 0x19, 0xFA] },
                { label: 'Open Drawer 2', code: '\\x1B\\x70\\x01\\x19\\xFA', desc: 'ESC p 1', bytes: [0x1B, 0x70, 0x01, 0x19, 0xFA] },
            ]
        },
        buzzer: {
            label: 'Buzzer',
            commands: [
                { label: 'Beep 1x', code: '\\x1B\\x42\\x01\\x02', desc: 'ESC B 1 2', bytes: [0x1B, 0x42, 0x01, 0x02] },
                { label: 'Beep 3x', code: '\\x1B\\x42\\x03\\x02', desc: 'ESC B 3 2', bytes: [0x1B, 0x42, 0x03, 0x02] },
                { label: 'Beep 5x', code: '\\x1B\\x42\\x05\\x02', desc: 'ESC B 5 2', bytes: [0x1B, 0x42, 0x05, 0x02] },
            ]
        },
        barcode: {
            label: 'Barcode',
            commands: [
                { label: 'Barcode Height 50', code: '\\x1D\\x68\\x32', desc: 'GS h 50', bytes: [0x1D, 0x68, 0x32] },
                { label: 'Barcode Height 100', code: '\\x1D\\x68\\x64', desc: 'GS h 100', bytes: [0x1D, 0x68, 0x64] },
                { label: 'Barcode Width 2', code: '\\x1D\\x77\\x02', desc: 'GS w 2', bytes: [0x1D, 0x77, 0x02] },
                { label: 'Barcode Width 3', code: '\\x1D\\x77\\x03', desc: 'GS w 3', bytes: [0x1D, 0x77, 0x03] },
                { label: 'HRI Below', code: '\\x1D\\x48\\x02', desc: 'GS H 2', bytes: [0x1D, 0x48, 0x02] },
                { label: 'HRI Off', code: '\\x1D\\x48\\x00', desc: 'GS H 0', bytes: [0x1D, 0x48, 0x00] },
            ]
        },
        qrcode: {
            label: 'QR Code',
            commands: [
                { label: 'QR Model 2', code: '\\x1D\\x28\\x6B\\x04\\x00\\x31\\x41\\x32\\x00', desc: 'GS ( k', bytes: [0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00] },
                { label: 'QR Size 4', code: '\\x1D\\x28\\x6B\\x03\\x00\\x31\\x43\\x04', desc: 'GS ( k', bytes: [0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, 0x04] },
                { label: 'QR Size 8', code: '\\x1D\\x28\\x6B\\x03\\x00\\x31\\x43\\x08', desc: 'GS ( k', bytes: [0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, 0x08] },
                { label: 'QR Error M', code: '\\x1D\\x28\\x6B\\x03\\x00\\x31\\x45\\x31', desc: 'GS ( k - Level M', bytes: [0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, 0x31] },
                { label: 'QR Print', code: '\\x1D\\x28\\x6B\\x03\\x00\\x31\\x51\\x30', desc: 'GS ( k', bytes: [0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30] },
            ]
        },
        status: {
            label: 'Status',
            commands: [
                { label: 'Request Status', code: '\\x10\\x04\\x01', desc: 'DLE EOT 1', bytes: [0x10, 0x04, 0x01] },
                { label: 'Paper Status', code: '\\x10\\x04\\x04', desc: 'DLE EOT 4', bytes: [0x10, 0x04, 0x04] },
            ]
        },
    };

    function formatBytes(bytes: number[]): string {
        return bytes.map(b => b.toString(16).padStart(2, '0').toUpperCase()).join(' ');
    }

    function insertCommand(cmd: {label: string, code: string, desc: string, bytes: number[]}) {
        if (editorView) {
            const pos = editorView.state.selection.main.head;
            editorView.dispatch({
                changes: { from: pos, insert: cmd.code },
                selection: { anchor: pos + cmd.code.length }
            });
            editorView.focus();
            
            // Log the command
            commandLog = [...commandLog, {
                name: cmd.label,
                bytes: cmd.desc,
                hex: formatBytes(cmd.bytes)
            }];
            
            // Keep only last 10 entries
            if (commandLog.length > 10) {
                commandLog = commandLog.slice(-10);
            }
        }
    }

    function clearLog() {
        commandLog = [];
    }

    onMount(() => {
        const updateListener = EditorView.updateListener.of((update) => {
            if (update.docChanged) {
                value = update.state.doc.toString();
                onchange(value);
            }
        });

        const state = EditorState.create({
            doc: value,
            extensions: [
                lineNumbers(),
                highlightActiveLineGutter(),
                highlightSpecialChars(),
                history(),
                drawSelection(),
                highlightActiveLine(),
                keymap.of([...defaultKeymap, ...historyKeymap]),
                oneDark,
                updateListener,
                EditorView.theme({
                    '&': {
                        height: '200px',
                        fontSize: '13px',
                    },
                    '.cm-scroller': {
                        overflow: 'auto',
                        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                    },
                    '.cm-content': {
                        minHeight: '180px',
                    },
                    '&.cm-focused': {
                        outline: 'none',
                    },
                }),
                EditorState.tabSize.of(4),
                EditorView.lineWrapping,
            ],
        });

        editorView = new EditorView({
            state,
            parent: editorContainer,
        });
    });

    onDestroy(() => {
        editorView?.destroy();
    });

    $effect(() => {
        if (editorView && editorView.state.doc.toString() !== value) {
            editorView.dispatch({
                changes: { from: 0, to: editorView.state.doc.length, insert: value }
            });
        }
    });
</script>

<div class="escpos-editor">
    <!-- Command Palette with Categories -->
    <div class="command-palette">
        <div class="category-tabs">
            {#each Object.entries(commandCategories) as [key, cat]}
                <button 
                    type="button"
                    class={`cat-tab ${selectedCategory === key ? 'active' : ''}`}
                    onclick={() => selectedCategory = key}
                >
                    {cat.label}
                </button>
            {/each}
        </div>
        <div class="command-buttons">
            {#each commandCategories[selectedCategory].commands as cmd}
                <button 
                    type="button"
                    class="cmd-btn"
                    title={`${cmd.desc} | Bytes: ${formatBytes(cmd.bytes)}`}
                    onclick={() => insertCommand(cmd)}
                >
                    {cmd.label}
                </button>
            {/each}
        </div>
    </div>

    <!-- Editor Container -->
    <div class="editor-wrapper" bind:this={editorContainer}></div>

    <!-- Command Log -->
    {#if commandLog.length > 0}
        <div class="command-log">
            <div class="log-header">
                <span class="log-title">Command Log</span>
                <button type="button" class="clear-log" onclick={clearLog}>Clear</button>
            </div>
            <div class="log-entries">
                {#each commandLog as entry, i}
                    <div class="log-entry">
                        <span class="entry-num">{i + 1}.</span>
                        <span class="entry-name">{entry.name}</span>
                        <span class="entry-desc">{entry.bytes}</span>
                        <code class="entry-hex">{entry.hex}</code>
                    </div>
                {/each}
            </div>
        </div>
    {/if}

    <!-- Help Text -->
    <div class="help-text">
        <span>Use <code>\x1B</code> for ESC, <code>\x1D</code> for GS, <code>\n</code> for newline</span>
    </div>
</div>

<style>
    .escpos-editor {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .command-palette {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .category-tabs {
        display: flex;
        flex-wrap: wrap;
        gap: 2px;
        padding-bottom: 6px;
        border-bottom: 1px solid #374151;
    }

    .cat-tab {
        padding: 4px 10px;
        font-size: 10px;
        text-transform: uppercase;
        font-weight: 600;
        background: transparent;
        border: none;
        border-radius: 4px;
        color: #6b7280;
        cursor: pointer;
        transition: all 0.15s ease;
    }

    .cat-tab:hover {
        color: #9ca3af;
        background: rgba(255, 255, 255, 0.05);
    }

    .cat-tab.active {
        color: #a5b4fc;
        background: rgba(99, 102, 241, 0.15);
    }

    .command-buttons {
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
    }

    .cmd-btn {
        padding: 4px 8px;
        font-size: 11px;
        background: rgba(99, 102, 241, 0.1);
        border: 1px solid rgba(99, 102, 241, 0.3);
        border-radius: 4px;
        color: #a5b4fc;
        cursor: pointer;
        transition: all 0.15s ease;
    }

    .cmd-btn:hover {
        background: rgba(99, 102, 241, 0.2);
        border-color: rgba(99, 102, 241, 0.5);
        color: #c7d2fe;
    }

    .editor-wrapper {
        border: 1px solid #374151;
        border-radius: 6px;
        overflow: hidden;
        background: #1f2937;
    }

    .command-log {
        background: #111827;
        border: 1px solid #374151;
        border-radius: 6px;
        padding: 8px;
        font-size: 11px;
    }

    .log-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 6px;
        padding-bottom: 6px;
        border-bottom: 1px solid #374151;
    }

    .log-title {
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: #9ca3af;
    }

    .clear-log {
        padding: 2px 8px;
        font-size: 10px;
        background: transparent;
        border: 1px solid #4b5563;
        border-radius: 3px;
        color: #6b7280;
        cursor: pointer;
    }

    .clear-log:hover {
        border-color: #ef4444;
        color: #ef4444;
    }

    .log-entries {
        display: flex;
        flex-direction: column;
        gap: 4px;
        max-height: 120px;
        overflow-y: auto;
    }

    .log-entry {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 4px 6px;
        background: rgba(255, 255, 255, 0.02);
        border-radius: 3px;
    }

    .entry-num {
        color: #4b5563;
        font-weight: 500;
        width: 20px;
    }

    .entry-name {
        color: #e5e7eb;
        font-weight: 500;
        min-width: 100px;
    }

    .entry-desc {
        color: #6b7280;
        flex: 1;
    }

    .entry-hex {
        background: rgba(16, 185, 129, 0.1);
        padding: 2px 6px;
        border-radius: 3px;
        font-family: monospace;
        color: #10b981;
        font-size: 10px;
    }

    .help-text {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 11px;
        color: #6b7280;
    }

    .help-text code {
        background: rgba(99, 102, 241, 0.1);
        padding: 1px 4px;
        border-radius: 3px;
        font-family: monospace;
        color: #a5b4fc;
    }
</style>

