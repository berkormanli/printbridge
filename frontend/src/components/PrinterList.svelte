<script lang="ts">
    let { printers = [], selectedPrinter = null, onselect } = $props();

    let windowsPrinters = $derived(printers.filter(p => p.device_type === 'Windows'));
    let usbPrinters = $derived(printers.filter(p => p.device_type !== 'Windows'));
</script>

<div class="space-y-6">
    {#if windowsPrinters.length > 0}
        <div>
            <h3 class="text-xs uppercase text-blue-400 font-bold tracking-wider mb-2 px-1">Windows Printers</h3>
            <div class="space-y-2">
                {#each windowsPrinters as p}
                    <div 
                        class={`
                            group p-3 rounded-lg cursor-pointer transition-all duration-200 border
                            ${selectedPrinter && selectedPrinter.vendor_id === p.vendor_id && selectedPrinter.product_id === p.product_id && selectedPrinter.product === p.product
                                ? 'bg-blue-500/20 border-blue-500 shadow-sm' 
                                : 'bg-gray-800/30 border-transparent hover:bg-gray-800 hover:border-gray-700'}
                        `}
                        onclick={() => onselect(p)}
                        onkeydown={(e) => e.key === 'Enter' && onselect(p)}
                        role="button"
                        tabindex="0"
                    >
                        <div class="flex justify-between items-start">
                            <div class="min-w-0 flex-1">
                                <div class="font-medium text-sm truncate text-gray-200 group-hover:text-white transition-colors">
                                    {p.product || 'Unknown Printer'}
                                </div>
                                <div class="text-xs text-gray-500 mt-1 flex items-center gap-2">
                                    <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-bold tracking-wider bg-blue-500/20 text-blue-400">
                                        Windows Spooler
                                    </span>
                                    <span>{p.manufacturer}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}

    {#if usbPrinters.length > 0}
        <div>
            <h3 class="text-xs uppercase text-purple-400 font-bold tracking-wider mb-2 px-1">USB Devices</h3>
            <div class="space-y-2">
                {#each usbPrinters as p}
                    <div 
                        class={`
                            group p-3 rounded-lg cursor-pointer transition-all duration-200 border
                            ${selectedPrinter && selectedPrinter.vendor_id === p.vendor_id && selectedPrinter.product_id === p.product_id
                                ? 'bg-purple-500/20 border-purple-500 shadow-sm' 
                                : 'bg-gray-800/30 border-transparent hover:bg-gray-800 hover:border-gray-700'}
                        `}
                        onclick={() => onselect(p)}
                        onkeydown={(e) => e.key === 'Enter' && onselect(p)}
                        role="button"
                        tabindex="0"
                    >
                        <div class="flex justify-between items-start">
                            <div class="min-w-0 flex-1">
                                <div class="font-medium text-sm truncate text-gray-200 group-hover:text-white transition-colors">
                                    {p.product || 'Unknown Device'}
                                </div>
                                <div class="text-xs text-gray-500 mt-1 flex items-center gap-2">
                                    <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-bold tracking-wider bg-purple-500/20 text-purple-400">
                                        Raw USB
                                    </span>
                                    <span>VID:{p.vendor_id?.toString(16).padStart(4, '0')} PID:{p.product_id?.toString(16).padStart(4, '0')}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}

    {#if printers.length === 0}
        <div class="text-center py-8 text-gray-500 text-sm">
            No printers found
        </div>
    {/if}
</div>
