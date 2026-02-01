<script lang="ts">
   // @ts-ignore
   import { PrintTest, SendRaw } from '../../wailsjs/go/main/App';
   import ESCPOSEditor from './ESCPOSEditor.svelte';

   let rawData = $state("");
   let quickString = $state("");

   async function runTest(type: string) {
       try {
           await PrintTest(type);
           console.log('Test print sent!');
       } catch (err) {
           console.error(err);
       }
   }

   async function sendRaw() {
       if (!rawData) return;
       try {
           await SendRaw(rawData);
           console.log('Raw data sent!');
           rawData = "";
       } catch (err) {
           console.error(err);
       }
   }

   async function sendQuickString() {
       if (!quickString) return;
       try {
           await SendRaw(quickString);
           console.log('Quick string sent!');
           quickString = "";
       } catch (err) {
           console.error(err);
       }
   }
</script>

<div class="space-y-6">
   <div class="grid grid-cols-2 gap-4">
       <button 
           onclick={() => runTest('simple')}
           class="p-4 bg-gray-800 hover:bg-gray-700 rounded-lg border border-gray-700 hover:border-primary/50 transition-all flex flex-col items-center gap-3 group"
       >
           <div class="w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center text-blue-400 group-hover:bg-blue-500 group-hover:text-white transition-colors">
               <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
           </div>
           <span class="font-medium text-sm">Simple Test</span>
       </button>

       <button 
           onclick={() => runTest('comprehensive')}
           class="p-4 bg-gray-800 hover:bg-gray-700 rounded-lg border border-gray-700 hover:border-purple-500/50 transition-all flex flex-col items-center gap-3 group"
       >
           <div class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-400 group-hover:bg-purple-500 group-hover:text-white transition-colors">
               <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"></path></svg>
           </div>
           <span class="font-medium text-sm">Full Report</span>
       </button>
   </div>

   <!-- Quick String Input -->
   <div class="pt-4 border-t border-gray-700/50">
       <label class="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
           Quick String
       </label>
       <div class="flex gap-2">
           <input 
               type="text"
               bind:value={quickString}
               placeholder="Enter text to send..."
               class="flex-1 bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-primary transition-colors"
               onkeydown={(e) => e.key === 'Enter' && sendQuickString()}
           />
           <button 
               onclick={sendQuickString}
               disabled={!quickString}
               class="bg-green-600 hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded text-sm font-medium transition-colors whitespace-nowrap"
           >
               Send String
           </button>
       </div>
   </div>

   <div class="pt-4 border-t border-gray-700/50">
       <label class="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
           Raw ESC/POS Editor
       </label>
       
       <ESCPOSEditor bind:value={rawData} />
       
       <div class="flex justify-end gap-2 mt-3">
           <button 
               onclick={() => rawData = ''}
               disabled={!rawData}
               class="px-4 py-2 rounded text-sm font-medium text-gray-400 hover:text-white transition-colors disabled:opacity-50"
           >
               Clear
           </button>
           <button 
               onclick={sendRaw}
               disabled={!rawData}
               class="bg-primary hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-2 rounded text-sm font-medium transition-colors"
           >
               Send Raw Data
           </button>
       </div>
   </div>
</div>

