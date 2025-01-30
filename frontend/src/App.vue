<script setup lang="ts">
import { ref } from 'vue'

const ws = ref<WebSocket | null>(null)
const isBroadcasting = ref(false)
const isListening = ref(false)

let mediaRecorder: MediaRecorder | null = null
let audioChunks: Blob[] = []
let audioContext: AudioContext | null = null
let source: AudioBufferSourceNode | null = null

// Start broadcasting
const startBroadcasting = async () => {
  if (!navigator.mediaDevices.getUserMedia) {
    alert("Your browser does not support the MediaStream API")
    return
  }

  ws.value = new WebSocket('ws://localhost:8080/ws?mode=broadcast')

  ws.value.onopen = () => {
    console.log("Conected to the server as a broadcaster")
    isBroadcasting.value = true
  }

  const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
  mediaRecorder = new MediaRecorder(stream)

  mediaRecorder.ondataavailable = (e) => {
    ws.value?.send(e.data)
  }

  mediaRecorder.start(50) // Send audio in 50ms chunks
}

const stopBroadcasting = () => {
  mediaRecorder?.stop()
  ws.value?.close()
  isBroadcasting.value = false
}

// Start listening (Opt-in)
const startListening = async () => {
  ws.value = new WebSocket('ws://localhost:8080/ws?mode=listen')

  ws.value.onmessage = event => {
    if (!audioContext) {
      audioContext = new AudioContext()
    }

    event.data.arrayBuffer().then((buffer: ArrayBuffer) => {
      if (audioContext) {
        audioContext.decodeAudioData(buffer, (decoded: AudioBuffer) => {
          source = audioContext!!.createBufferSource()
          if (source) {
            source.buffer = decoded
            source.connect(audioContext!!.destination)
            source.start()
          }
        })
      }
    })
  }

  isListening.value = true
}

const stopListening = () => {
  if (source) source.stop()
  if (ws.value) ws.value.close()
  isListening.value = false
}
</script>

<template>
  <div class="container text-center mt-5">
    <h1 class="mb-4 text-primary">🎤 EchoLive - Karaoke Broadcast</h1>

    <div class="d-flex justify-content-center gap-3">
      <!-- Start Broadcast Button -->
      <button 
        v-if="!isBroadcasting" 
        @click="startBroadcasting" 
        class="btn btn-success">
        🎙 Start Broadcasting
      </button>

      <!-- Stop Broadcast Button -->
      <button 
        v-if="isBroadcasting" 
        @click="stopBroadcasting" 
        class="btn btn-danger">
        ⛔ Stop Broadcasting
      </button>
    </div>

    <div class="mt-4 d-flex justify-content-center gap-3">
      <!-- Start Listening Button -->
      <button 
        v-if="!isListening" 
        @click="startListening" 
        class="btn btn-primary">
        🎧 Listen (Opt-in)
      </button>

      <!-- Stop Listening Button -->
      <button 
        v-if="isListening" 
        @click="stopListening" 
        class="btn btn-warning">
        🔇 Stop Listening
      </button>
    </div>
  </div>
</template>

<style scoped>
.container {
  max-width: 600px;
}
</style>