window.JanuszAudio = {
    musicAudio: null,
    voiceAudio: null,
    sfxAudio: null,
    musicBaseVol: 0.5,
    voiceBaseVol: 0.5,
    sfxBaseVol: 0.5,
    duckingLevel: 0.15,
    isVoiceMuted: false,
    
    setMusicVol: function(vol) {
        this.musicBaseVol = parseFloat(vol);
        if (this.musicAudio) {
            if (this.voiceAudio && !this.voiceAudio.paused && !this.voiceAudio.ended && !this.isVoiceMuted) {
                this.musicAudio.volume = this.duckingLevel;
            } else {
                this.musicAudio.volume = this.musicBaseVol;
            }
        }
    },
    // ALIASY dla starego silnika
    setSfxVol: function(vol) { this.setVoiceVol(vol); },
    setVoiceVolume: function(vol) { this.setVoiceVol(vol); },
    
    setVoiceVol: function(vol) {
        this.voiceBaseVol = parseFloat(vol);
        this.sfxBaseVol = parseFloat(vol);
        this.isVoiceMuted = vol == 0;
        if (this.voiceAudio) this.voiceAudio.volume = this.voiceBaseVol;
        if (this.sfxAudio) this.sfxAudio.volume = this.sfxBaseVol;
        if (this.isVoiceMuted && this.musicAudio) this.musicAudio.volume = this.musicBaseVol;
        else if (!this.isVoiceMuted && this.voiceAudio && !this.voiceAudio.paused && !this.voiceAudio.ended) {
            if (this.musicAudio) this.musicAudio.volume = this.duckingLevel;
        }
    },
    
    startMusic: function(src) {
        if (!this.musicAudio) {
            this.musicAudio = new Audio(src);
            this.musicAudio.loop = true;
            this.musicAudio.volume = this.musicBaseVol;
            this.musicAudio.play().catch(e => console.log("Muzyka czeka na klik:", e));
        } else {
            this.musicAudio.src = src;
            this.musicAudio.play().catch(()=>{});
        }
    },
    playMusic: function(src) { this.startMusic(src); },
    
    playVoice: function(src) {
        if (this.voiceAudio) { this.voiceAudio.pause(); this.voiceAudio.currentTime = 0; }
        this.voiceAudio = new Audio(src);
        this.voiceAudio.volume = this.voiceBaseVol;
        this.voiceAudio.addEventListener('play', () => {
            if (!this.isVoiceMuted && this.musicAudio && !this.musicAudio.paused) this.musicAudio.volume = this.duckingLevel;
        });
        const restoreMusic = () => { if (this.musicAudio && !this.musicAudio.paused) this.musicAudio.volume = this.musicBaseVol; };
        this.voiceAudio.addEventListener('ended', restoreMusic);
        this.voiceAudio.addEventListener('pause', restoreMusic);
        if (!this.isVoiceMuted) this.voiceAudio.play().catch(e => console.log("Voice error:", e));
    },
    playSfx: function(src) { this.playVoice(src); },
    playSound: function(src) { this.playVoice(src); },
    stopMusic: function() { if(this.musicAudio) this.musicAudio.pause(); }
};