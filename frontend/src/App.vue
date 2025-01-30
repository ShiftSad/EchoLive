<script setup lang="ts">
import { ref } from 'vue'

const ws = ref<WebSocket | null>(null)
const isBroadcasting = ref(false)
const isListening = ref(false)

let mediaRecorder: MediaRecorder | null = null
let audioChunks: Blob[] = []
let audioContext: AudioContext | null = null
let source: MediaStreamAudioSourceNode | null = null

// Start broadcasting
const startBroadcasting = async () => {
  if (!navigator.mediaDevices.getUserMedia) {
    alert("Your browser does not support the MediaStream API")
    return
  }

  ws.value = new WebSocket('ws://localhost:8000/ws?mode=broadcast')
}
</script>