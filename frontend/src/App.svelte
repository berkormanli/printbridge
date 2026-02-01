<script lang="ts">
  import { onMount } from 'svelte';
  // @ts-ignore
  import { GetPrinters, SelectPrinter, PrintTest, SendRaw, CheckServiceStatus, GetConnectionStatus } from '../wailsjs/go/main/App.js';
  import PrinterList from './components/PrinterList.svelte';
  import TestZone from './components/TestZone.svelte';

  let printers = $state([]);
  let selectedPrinter = $state(null);
  let serviceOnline = $state(false);
  let printerConnected = $state(false);
  let status = $state("Checking service...");

  async function checkService() {
    try {
      serviceOnline = await CheckServiceStatus();
      if (serviceOnline) {
        printerConnected = await GetConnectionStatus();
        if (printerConnected) {
          status = "Service Online • Printer Connected";
        } else {
          status = "Service Online • Printer Disconnected";
        }
      } else {
        status = "Service Offline";
        printers = [];
      }
    } catch (err) {
      console.error(err);
      serviceOnline = false;
      status = "Service Offline";
    }
  }

  async function refreshPrinters() {
    if (!serviceOnline) {
      await checkService();
      if (!serviceOnline) return;
    }
    try {
      printers = await GetPrinters();
      console.log('Printers refreshed');
    } catch (err) {
      console.error(err);
    }
  }

  async function onSelect(printer: any) {
    try {
      if (!printer.device_type) printer.device_type = "USB";
      await SelectPrinter(printer.product, printer.device_type, printer.vendor_id || 0, printer.product_id || 0);
      selectedPrinter = printer;
      status = "Connected to " + printer.product;
      console.log(`Selected ${printer.product}`);
    } catch (err) {
      console.error(err);
      status = "Error selecting printer";
    }
  }

  onMount(() => {
    checkService().then(() => {
      if (serviceOnline) refreshPrinters();
    });
    
    // Periodic status check
    const interval = setInterval(checkService, 5000);
    return () => clearInterval(interval);
  });
</script>

<div class="flex h-screen bg-app-bg text-white overflow-hidden font-sans">
  
  <!-- Sidebar -->
  <aside class="w-80 bg-card-bg border-r border-gray-700 flex flex-col">
    <div class="p-6 border-b border-gray-700 flex items-center gap-3">
      <div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center font-bold text-xl">P</div>
      <h1 class="text-xl font-bold tracking-tight">PrintBridge</h1>
    </div>

    <div class="flex-1 overflow-y-auto p-4">
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-xs uppercase text-gray-400 font-semibold tracking-wider">Printers</h2>
        <button onclick={refreshPrinters} class="text-xs text-primary hover:text-white transition-colors">
          Refresh
        </button>
      </div>

      <PrinterList {printers} {selectedPrinter} onselect={onSelect} />
    </div>

    <div class="p-4 bg-gray-900/50 text-xs text-center text-gray-500">
      v1.0.0
    </div>
  </aside>

  <!-- Main Content -->
  <main class="flex-1 flex flex-col min-w-0">
    <!-- Header -->
    <header class="h-16 border-b border-gray-700 flex items-center justify-between px-8 bg-card-bg/50 backdrop-blur-sm z-10">
      <div class="flex items-center gap-2">
        <div class={`w-2 h-2 rounded-full ${!serviceOnline ? 'bg-red-500' : printerConnected ? 'bg-green-500' : 'bg-yellow-500'}`}></div>
        <span class="text-sm font-medium text-gray-300">{status}</span>
      </div>
      <div>
      </div>
    </header>

    <!-- Content -->
    <div class="flex-1 p-8 overflow-y-auto">
      {#if selectedPrinter}
        <div class="max-w-4xl mx-auto space-y-6">
          <div class="bg-card-bg rounded-xl p-6 shadow-lg border border-gray-700/50">
            <h2 class="text-lg font-semibold mb-1">{selectedPrinter.product}</h2>
            <p class="text-sm text-gray-400 mb-6">{selectedPrinter.manufacturer || 'Unknown Manufacturer'} • {selectedPrinter.device_type}</p>
            
            <TestZone />
          </div>
        </div>
      {:else}
        <div class="h-full flex flex-col items-center justify-center text-gray-500 opacity-60">
          <svg class="w-16 h-16 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"></path></svg>
          <p class="text-lg">Select a printer to start</p>
        </div>
      {/if}
    </div>
  </main>
</div>

<style>
  :global(body) {
    background-color: #1b2636;
  }
</style>
