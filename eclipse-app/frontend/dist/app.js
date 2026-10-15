// Castlevania Eclipse Collection - Clean Frontend Controller

const DEFAULT_LIBRARY = [
  {
    id: "rondo",
    title: "Castlevania: Rondo of Blood",
    original_system: "PC Engine Super CD-ROM²",
    release_year: "1993",
    developer: "Konami",
    synopsis: "Richter Belmont sets out with the legendary whip Vampire Killer to rescue his beloved Annette and young Maria Renard from the resurrected Lord of Darkness.",
    supported_regions: ["US", "JP"],
    default_region: "US",
    theme_track: "assets/audio/rondo_theme.mp3"
  },
  {
    id: "sotn",
    title: "Castlevania: Symphony of the Night",
    original_system: "PlayStation / PC Recomp",
    release_year: "1997",
    developer: "Konami Computer Entertainment Tokyo",
    synopsis: "Dracula's castle rises once more, but Richter Belmont has vanished. Awoken from his eternal slumber, Alucard enters the inverted realm to end his father's curse.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/sotn_theme.mp3"
  },
  {
    id: "loi",
    title: "Castlevania: Lament of Innocence",
    original_system: "PlayStation 2",
    release_year: "2003",
    developer: "Konami Computer Entertainment Tokyo",
    synopsis: "In the 11th century, knight Leon Belmont renounces his title to enter the forest of eternal night and save his betrothed Sara from the vampire Walter Bernhard—unveiling the origin of the whip.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/loi_theme.mp3"
  },
  {
    id: "cod",
    title: "Castlevania: Curse of Darkness",
    original_system: "PlayStation 2",
    release_year: "2005",
    developer: "Konami Digital Entertainment",
    synopsis: "Three years after Trevor Belmont defeated Dracula, Hector, a former Devil Forgemaster seeking redemption, hunts his former comrade Isaac across a curse-ravaged Europe.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/cod_theme.mp3"
  },
  {
    id: "lod",
    title: "Castlevania: Legacy of Darkness",
    original_system: "Nintendo 64",
    release_year: "1999",
    developer: "Konami Computer Entertainment Kobe",
    synopsis: "The definitive 3D gothic quest on Nintendo 64. Play as Cornell the Man-Beast, Henry the Knight, Reinhardt Schneider, or Carrie Fernandez to rescue young Ada from the dark ritual.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/lod_theme.mp3"
  },
  {
    id: "cv64",
    title: "Castlevania 64",
    original_system: "Nintendo 64",
    release_year: "1999",
    developer: "Konami Computer Entertainment Kobe",
    synopsis: "Castlevania's first journey into three dimensions. Take control of whip-wielding Reinhardt Schneider or young magical prodigy Carrie Fernandez as they ascend Dracula's castle.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/cv64_theme.mp3"
  },
  {
    id: "rebirth",
    title: "Castlevania: The Adventure ReBirth",
    original_system: "WiiWare (Developed by M2)",
    release_year: "2009",
    developer: "M2 / Konami",
    synopsis: "M2's acclaimed remake of the classic adventure. Guide Christopher Belmont through punishing traps, branched routes, and classic gothic action with timeless arranged tracks.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/rebirth_theme.mp3"
  }
];

class EclipseLauncher {
  constructor() {
    this.games = DEFAULT_LIBRARY;
    this.currentIndex = 0;
    this.selectedRegions = {};
    this.audioEnabled = true;
    this.audioCtx = null;
    this.gamepadTimestamp = 0;

    // DOM Elements
    this.bgLayer = document.getElementById('bgLayer');
    this.systemBadge = document.getElementById('systemBadge');
    this.yearBadge = document.getElementById('yearBadge');
    this.devBadge = document.getElementById('devBadge');
    this.gameTitle = document.getElementById('gameTitle');
    this.gameSynopsis = document.getElementById('gameSynopsis');
    this.regionButtons = document.getElementById('regionButtons');
    this.launchBtn = document.getElementById('launchBtn');
    this.statusIndicator = document.getElementById('statusIndicator');
    this.statusText = document.getElementById('statusText');
    this.missingHint = document.getElementById('missingHint');
    this.missingHintText = document.getElementById('missingHintText');
    this.boxartImg = document.getElementById('boxartImg');
    this.boxartFallback = document.getElementById('boxartFallback');
    this.fallbackTitle = document.getElementById('fallbackTitle');
    this.fallbackSystem = document.getElementById('fallbackSystem');
    this.boxartRegionBadge = document.getElementById('boxartRegionBadge');
    this.carouselTrack = document.getElementById('carouselTrack');
    this.gameCounter = document.getElementById('gameCounter');
    this.fullscreenBtn = document.getElementById('fullscreenBtn');
    this.audioToggleBtn = document.getElementById('audioToggleBtn');
    this.audioLabel = document.getElementById('audioLabel');
    this.collectionLogo = document.getElementById('collectionLogo');
    this.collectionText = document.getElementById('collectionText');

    this.init();
  }

  async init() {
    this.checkCollectionLogo();
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const backendGames = await window.go.main.App.GetLibrary();
        if (backendGames && backendGames.length > 0) {
          this.games = backendGames;
        }
      } catch (e) {
        console.warn('Backend library unavailable, using local library');
      }
    }

    this.games.forEach(g => {
      this.selectedRegions[g.id] = g.default_region || (g.supported_regions ? g.supported_regions[0] : 'US');
    });

    this.renderCarousel();
    this.updateActiveGame();
    this.bindEvents();
    this.startGamepadLoop();
    this.initAudio();
  }

  checkCollectionLogo() {
    if (!this.collectionLogo) return;
    const img = new Image();
    img.onload = () => {
      this.collectionLogo.style.display = 'block';
      if (this.collectionText) this.collectionText.style.display = 'none';
    };
    img.onerror = () => {
      this.collectionLogo.style.display = 'none';
      if (this.collectionText) this.collectionText.style.display = 'flex';
    };
    img.src = 'assets/images/collection_logo.png';
  }

  initAudio() {
    try {
      const AudioContext = window.AudioContext || window.webkitAudioContext;
      if (AudioContext) {
        this.audioCtx = new AudioContext();
      }
    } catch (e) {}
  }

  playSfx(type) {
    if (!this.audioEnabled || !this.audioCtx) return;
    if (this.audioCtx.state === 'suspended') {
      this.audioCtx.resume();
    }

    const osc = this.audioCtx.createOscillator();
    const gain = this.audioCtx.createGain();
    const now = this.audioCtx.currentTime;

    osc.connect(gain);
    gain.connect(this.audioCtx.destination);

    if (type === 'cursor') {
      osc.type = 'triangle';
      osc.frequency.setValueAtTime(320, now);
      osc.frequency.exponentialRampToValueAtTime(140, now + 0.05);
      gain.gain.setValueAtTime(0.12, now);
      gain.gain.exponentialRampToValueAtTime(0.01, now + 0.05);
      osc.start(now);
      osc.stop(now + 0.05);
    } else if (type === 'confirm') {
      osc.type = 'sine';
      osc.frequency.setValueAtTime(523.25, now);
      osc.frequency.setValueAtTime(659.25, now + 0.05);
      gain.gain.setValueAtTime(0.18, now);
      gain.gain.exponentialRampToValueAtTime(0.01, now + 0.3);
      osc.start(now);
      osc.stop(now + 0.3);
    } else if (type === 'launch') {
      osc.type = 'sawtooth';
      osc.frequency.setValueAtTime(220, now);
      osc.frequency.exponentialRampToValueAtTime(880, now + 0.15);
      gain.gain.setValueAtTime(0.22, now);
      gain.gain.exponentialRampToValueAtTime(0.01, now + 0.4);
      osc.start(now);
      osc.stop(now + 0.4);
    }
  }

  renderCarousel() {
    this.carouselTrack.innerHTML = '';
    this.games.forEach((game, index) => {
      const card = document.createElement('div');
      card.className = `carousel-card ${index === this.currentIndex ? 'active' : ''}`;
      card.setAttribute('data-index', index);
      card.innerHTML = `
        <div class="carousel-title">${game.title}</div>
        <div class="carousel-sub">${game.original_system || ''}</div>
        <div class="carousel-badge">${game.release_year || ''}</div>
      `;
      card.addEventListener('click', () => {
        this.selectGame(index);
      });
      this.carouselTrack.appendChild(card);
    });
  }

  selectGame(index) {
    if (index === this.currentIndex) return;
    this.currentIndex = (index + this.games.length) % this.games.length;
    this.playSfx('cursor');
    this.updateActiveGame();
    this.updateCarouselSelection();
  }

  updateCarouselSelection() {
    const cards = this.carouselTrack.querySelectorAll('.carousel-card');
    cards.forEach((card, idx) => {
      card.classList.toggle('active', idx === this.currentIndex);
    });

    const activeCard = cards[this.currentIndex];
    if (activeCard) {
      const offset = activeCard.offsetLeft - (this.carouselTrack.parentElement.clientWidth / 2) + (activeCard.clientWidth / 2);
      this.carouselTrack.style.transform = `translateX(${-Math.max(0, offset)}px)`;
    }
    this.gameCounter.textContent = `${this.currentIndex + 1} / ${this.games.length}`;
  }

  async updateActiveGame() {
    const game = this.games[this.currentIndex];
    if (!game) return;

    const currentRegion = this.selectedRegions[game.id] || game.default_region || 'US';

    this.gameTitle.textContent = game.title;
    this.systemBadge.textContent = (game.original_system || 'KONAMI').toUpperCase();
    this.yearBadge.textContent = game.release_year || '';
    this.devBadge.textContent = (game.developer || 'KONAMI').toUpperCase();
    this.gameSynopsis.textContent = game.synopsis || '';
    this.fallbackTitle.textContent = game.title.replace('Castlevania: ', '');
    this.fallbackSystem.textContent = game.original_system || '';
    this.boxartRegionBadge.textContent = currentRegion;

    this.renderRegionButtons(game, currentRegion);
    this.loadCustomAssets(game, currentRegion);
    this.checkStatus(game, currentRegion);
  }

  renderRegionButtons(game, currentRegion) {
    this.regionButtons.innerHTML = '';
    const regions = game.supported_regions || ['US', 'JP', 'EU'];
    const labels = { US: 'USA', JP: 'JPN', EU: 'EUR' };

    regions.forEach(reg => {
      const btn = document.createElement('button');
      btn.className = `region-btn ${reg === currentRegion ? 'active' : ''}`;
      btn.textContent = labels[reg] || reg;
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.setRegion(game.id, reg);
      });
      this.regionButtons.appendChild(btn);
    });
  }

  setRegion(gameId, region) {
    if (this.selectedRegions[gameId] === region) return;
    this.selectedRegions[gameId] = region;
    this.playSfx('confirm');
    this.updateActiveGame();
  }

  cycleRegion() {
    const game = this.games[this.currentIndex];
    const regions = game.supported_regions || ['US', 'JP', 'EU'];
    const current = this.selectedRegions[game.id] || regions[0];
    const nextIdx = (regions.indexOf(current) + 1) % regions.length;
    this.setRegion(game.id, regions[nextIdx]);
  }

  loadCustomAssets(game, region) {
    const regionLower = region.toLowerCase();
    const boxartPath = `assets/games/${game.id}/box_${regionLower}.png`;
    const bgPath = `assets/games/${game.id}/background.png`;

    const testImg = new Image();
    testImg.onload = () => {
      this.boxartImg.src = boxartPath;
      this.boxartImg.style.display = 'block';
      this.boxartFallback.style.display = 'none';
    };
    testImg.onerror = () => {
      this.boxartImg.style.display = 'none';
      this.boxartFallback.style.display = 'flex';
    };
    testImg.src = boxartPath;

    const testBg = new Image();
    testBg.onload = () => {
      this.bgLayer.style.backgroundImage = `url('${bgPath}')`;
    };
    testBg.onerror = () => {
      this.bgLayer.style.backgroundImage = `url('assets/images/background.png')`;
    };
    testBg.src = bgPath;
  }

  async checkStatus(game, region) {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const status = await window.go.main.App.CheckGameStatus(game.id, region);
        if (status.available) {
          this.statusIndicator.querySelector('.status-dot').className = 'status-dot ready';
          this.statusText.textContent = 'READY TO PLAY';
          this.missingHint.style.display = 'none';
        } else {
          this.statusIndicator.querySelector('.status-dot').className = 'status-dot missing';
          this.statusText.textContent = 'ROM REQUIRED';
          this.missingHint.style.display = 'flex';
          this.missingHintText.textContent = `Place ROM in: ${status.missing_file || 'games/' + game.id + '/'}`;
        }
      } catch (e) {}
    }
  }

  async launchCurrentGame() {
    const game = this.games[this.currentIndex];
    const region = this.selectedRegions[game.id] || 'US';
    this.playSfx('launch');

    if (window.go && window.go.main && window.go.main.App) {
      try {
        await window.go.main.App.LaunchGame(game.id, region);
      } catch (err) {
        alert(err.message || 'Error launching game. Check ROM path.');
      }
    }
  }

  toggleFullscreen() {
    if (window.go && window.go.main && window.go.main.App) {
      window.go.main.App.ToggleFullscreen();
    } else {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().catch(() => {});
      } else {
        document.exitFullscreen().catch(() => {});
      }
    }
  }

  toggleAudio() {
    this.audioEnabled = !this.audioEnabled;
    this.audioLabel.textContent = this.audioEnabled ? 'AUDIO ON' : 'AUDIO OFF';
    if (!this.audioEnabled) {
      this.bgmPlayer.pause();
    }
  }

  bindEvents() {
    this.launchBtn.addEventListener('click', () => this.launchCurrentGame());
    this.fullscreenBtn.addEventListener('click', () => this.toggleFullscreen());
    this.audioToggleBtn.addEventListener('click', () => this.toggleAudio());

    window.addEventListener('keydown', (e) => {
      switch (e.key) {
        case 'ArrowLeft':
          this.selectGame(this.currentIndex - 1);
          break;
        case 'ArrowRight':
          this.selectGame(this.currentIndex + 1);
          break;
        case 'ArrowUp':
        case 'ArrowDown':
        case 'r':
        case 'R':
          this.cycleRegion();
          break;
        case 'Enter':
        case ' ':
          this.launchCurrentGame();
          break;
        case 'F11':
          e.preventDefault();
          this.toggleFullscreen();
          break;
        case 'm':
        case 'M':
          this.toggleAudio();
          break;
      }
    });
  }

  startGamepadLoop() {
    const poll = () => {
      const gamepads = navigator.getGamepads ? navigator.getGamepads() : [];
      const gp = gamepads[0] || gamepads[1];

      if (gp) {
        const now = performance.now();
        if (now - this.gamepadTimestamp > 220) {
          const axisX = gp.axes[0];
          const dpadLeft = gp.buttons[14] && gp.buttons[14].pressed;
          const dpadRight = gp.buttons[15] && gp.buttons[15].pressed;
          const dpadUp = gp.buttons[12] && gp.buttons[12].pressed;
          const dpadDown = gp.buttons[13] && gp.buttons[13].pressed;

          if (dpadLeft || axisX < -0.5) {
            this.selectGame(this.currentIndex - 1);
            this.gamepadTimestamp = now;
          } else if (dpadRight || axisX > 0.5) {
            this.selectGame(this.currentIndex + 1);
            this.gamepadTimestamp = now;
          } else if (dpadUp || dpadDown || (gp.buttons[3] && gp.buttons[3].pressed)) {
            this.cycleRegion();
            this.gamepadTimestamp = now;
          } else if (gp.buttons[0] && gp.buttons[0].pressed) {
            this.launchCurrentGame();
            this.gamepadTimestamp = now + 400;
          } else if (gp.buttons[9] && gp.buttons[9].pressed) {
            this.toggleFullscreen();
            this.gamepadTimestamp = now + 400;
          }
        }
      }
      requestAnimationFrame(poll);
    };
    requestAnimationFrame(poll);
  }
}

window.addEventListener('DOMContentLoaded', () => {
  new EclipseLauncher();
});
