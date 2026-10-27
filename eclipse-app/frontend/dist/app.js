// Castlevania Eclipse Collection - Frontend Controller (v1.0.4)
// Supports Start Screen, Audio, Isolated Emulator Container Manager, and In-Launcher Graphics/Audio Settings

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
    id: "rebirth",
    title: "Castlevania: The Adventure ReBirth",
    original_system: "WiiWare",
    release_year: "2009",
    developer: "Konami",
    synopsis: "M2's acclaimed remake of the classic adventure. Guide Christopher Belmont through punishing traps, branched routes, and classic gothic action with timeless arranged tracks.",
    supported_regions: ["US", "JP", "EU"],
    default_region: "US",
    theme_track: "assets/audio/rebirth_theme.mp3"
  }
  /* In Reserve (Can be re-enabled at any time):
  ,{
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
  }
  */
];

class EclipseLauncher {
  constructor() {
    this.games = DEFAULT_LIBRARY;
    this.currentIndex = 0;
    this.selectedRegions = {};
    this.audioEnabled = true;
    this.uiSfxEnabled = true;
    this.menuVolume = 50;
    this.audioCtx = null;
    this.gamepadTimestamp = 0;
    this.inStartScreen = true;
    this.inSettings = false;

    this.settings = {
      display_mode: "fullscreen",
      internal_resolution: 3,
      aspect_ratio: "4:3",
      renderer: "auto",
      shader_filter: "sharp",
      vsync: true,
      show_fps: false,
      menu_music: true,
      menu_volume: 50,
      ui_sfx: true,
      audio_sync: true,
      custom_rom_dir: ""
    };

    // DOM Elements - Start Screen
    this.startScreen = document.getElementById('startScreen');
    this.startLogo = document.getElementById('startLogo');
    this.startLogoFallback = document.getElementById('startLogoFallback');
    this.startBg = document.getElementById('startBg');

    // DOM Elements - Main Menu
    this.bgLayer = document.getElementById('bgLayer');
    this.systemBadge = document.getElementById('systemBadge');
    this.yearBadge = document.getElementById('yearBadge');
    this.gameTitle = document.getElementById('gameTitle');
    this.gameLogoImg = document.getElementById('gameLogoImg');
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
    this.settingsBtn = document.getElementById('settingsBtn');
    this.fullscreenBtn = document.getElementById('fullscreenBtn');
    this.audioToggleBtn = document.getElementById('audioToggleBtn');
    this.audioLabel = document.getElementById('audioLabel');
    this.collectionLogo = document.getElementById('collectionLogo');
    this.collectionText = document.getElementById('collectionText');
    this.bgmPlayer = document.getElementById('bgmPlayer');

    // DOM Elements - Settings Modal
    this.settingsOverlay = document.getElementById('settingsOverlay');
    this.settingsCloseBtn = document.getElementById('settingsCloseBtn');
    this.saveSettingsBtn = document.getElementById('saveSettingsBtn');
    this.settingsSaveStatus = document.getElementById('settingsSaveStatus');
    this.provisionAllBtn = document.getElementById('provisionAllBtn');
    this.scanRomsBtn = document.getElementById('scanRomsBtn');
    this.provisionProgress = document.getElementById('provisionProgress');
    this.provisionProgressText = document.getElementById('provisionProgressText');
    this.containersList = document.getElementById('containersList');

    // Settings Inputs
    this.optDisplayMode = document.getElementById('optDisplayMode');
    this.optResolution = document.getElementById('optResolution');
    this.optAspectRatio = document.getElementById('optAspectRatio');
    this.optRenderer = document.getElementById('optRenderer');
    this.optShaderFilter = document.getElementById('optShaderFilter');
    this.optVsync = document.getElementById('optVsync');
    this.optShowFps = document.getElementById('optShowFps');
    this.optMenuMusic = document.getElementById('optMenuMusic');
    this.optMenuVolume = document.getElementById('optMenuVolume');
    this.volValueLabel = document.getElementById('volValueLabel');
    this.optUiSfx = document.getElementById('optUiSfx');
    this.optAudioSync = document.getElementById('optAudioSync');

    // DOM Elements - Dialog Modal
    this.dialogOverlay = document.getElementById('dialogOverlay');
    this.dialogTitle = document.getElementById('dialogTitle');
    this.dialogMessage = document.getElementById('dialogMessage');
    this.dialogTip = document.getElementById('dialogTip');
    this.dialogActionBtn = document.getElementById('dialogActionBtn');
    this.dialogCloseBtn = document.getElementById('dialogCloseBtn');

    this.init();
  }

  async init() {
    this.checkCollectionLogo();
    this.checkTitleBackground();

    if (window.go && window.go.main && window.go.main.App) {
      try {
        const backendGames = await window.go.main.App.GetLibrary();
        if (backendGames && backendGames.length > 0) {
          this.games = backendGames;
        }
      } catch (e) {
        console.warn('Backend library unavailable, using local library');
      }

      try {
        const loadedSettings = await window.go.main.App.GetSettings();
        if (loadedSettings) {
          this.settings = loadedSettings;
          this.applySettingsToFrontendState();
        }
      } catch (e) {}
    }

    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('container:progress', (msg) => {
        if (this.provisionProgress && this.provisionProgressText) {
          this.provisionProgress.style.display = 'block';
          this.provisionProgressText.textContent = msg;
        }
      });
      window.runtime.EventsOn('game:started', () => {
        if (this.bgmPlayer) this.bgmPlayer.pause();
      });
      window.runtime.EventsOn('game:stopped', () => {
        if (this.audioEnabled && !this.inStartScreen) {
          this.startBgm();
        }
      });
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

  applySettingsToFrontendState() {
    this.audioEnabled = this.settings.menu_music !== false;
    this.uiSfxEnabled = this.settings.ui_sfx !== false;
    this.menuVolume = typeof this.settings.menu_volume === 'number' ? this.settings.menu_volume : 50;

    if (this.audioLabel) {
      this.audioLabel.textContent = this.audioEnabled ? 'AUDIO ON' : 'AUDIO OFF';
    }
    if (this.bgmPlayer) {
      this.bgmPlayer.volume = Math.max(0, Math.min(1, this.menuVolume / 100));
      if (!this.audioEnabled) {
        this.bgmPlayer.pause();
      } else if (!this.inStartScreen) {
        this.startBgm();
      }
    }

    // Populate Settings Form Controls
    if (this.optDisplayMode) this.optDisplayMode.value = this.settings.display_mode || 'fullscreen';
    if (this.optResolution) this.optResolution.value = String(this.settings.internal_resolution || 3);
    if (this.optAspectRatio) this.optAspectRatio.value = this.settings.aspect_ratio || '4:3';
    if (this.optRenderer) this.optRenderer.value = this.settings.renderer || 'auto';
    if (this.optShaderFilter) this.optShaderFilter.value = this.settings.shader_filter || 'sharp';
    if (this.optVsync) this.optVsync.value = String(this.settings.vsync !== false);
    if (this.optShowFps) this.optShowFps.value = String(!!this.settings.show_fps);
    if (this.optMenuMusic) this.optMenuMusic.value = String(this.settings.menu_music !== false);
    if (this.optMenuVolume) this.optMenuVolume.value = String(this.menuVolume);
    if (this.volValueLabel) this.volValueLabel.textContent = `${this.menuVolume}%`;
    if (this.optUiSfx) this.optUiSfx.value = String(this.settings.ui_sfx !== false);
    if (this.optAudioSync) this.optAudioSync.value = String(this.settings.audio_sync !== false);
  }

  readSettingsFromForm() {
    return {
      display_mode: this.optDisplayMode ? this.optDisplayMode.value : 'fullscreen',
      internal_resolution: this.optResolution ? parseInt(this.optResolution.value, 10) : 3,
      aspect_ratio: this.optAspectRatio ? this.optAspectRatio.value : '4:3',
      renderer: this.optRenderer ? this.optRenderer.value : 'auto',
      shader_filter: this.optShaderFilter ? this.optShaderFilter.value : 'sharp',
      vsync: this.optVsync ? this.optVsync.value === 'true' : true,
      show_fps: this.optShowFps ? this.optShowFps.value === 'true' : false,
      menu_music: this.optMenuMusic ? this.optMenuMusic.value === 'true' : true,
      menu_volume: this.optMenuVolume ? parseInt(this.optMenuVolume.value, 10) : 50,
      ui_sfx: this.optUiSfx ? this.optUiSfx.value === 'true' : true,
      audio_sync: this.optAudioSync ? this.optAudioSync.value === 'true' : true,
      custom_rom_dir: this.settings.custom_rom_dir || ''
    };
  }

  async saveSettings(closeAfter = false) {
    this.settings = this.readSettingsFromForm();
    this.applySettingsToFrontendState();
    this.playSfx('confirm');

    if (window.go && window.go.main && window.go.main.App) {
      try {
        await window.go.main.App.SaveSettings(this.settings);
        if (this.settingsSaveStatus) {
          this.settingsSaveStatus.textContent = 'Settings saved and injected into all isolated emulator containers.';
        }
      } catch (err) {
        if (this.settingsSaveStatus) {
          this.settingsSaveStatus.textContent = 'Error saving settings: ' + String(err);
        }
      }
    }

    if (closeAfter) {
      this.closeSettings();
    }
  }

  openSettings(initialTab = null) {
    this.inSettings = true;
    this.applySettingsToFrontendState();
    if (initialTab) {
      this.switchSettingsTab(initialTab);
    }
    this.refreshContainersStatus();
    if (this.settingsOverlay) {
      this.settingsOverlay.style.display = 'flex';
    }
    this.playSfx('confirm');
  }

  closeSettings() {
    this.inSettings = false;
    if (this.settingsOverlay) {
      this.settingsOverlay.style.display = 'none';
    }
    this.updateActiveGame();
  }

  switchSettingsTab(tabId) {
    const tabs = document.querySelectorAll('.settings-tab');
    const panels = document.querySelectorAll('.settings-panel');
    tabs.forEach(t => t.classList.toggle('active', t.getAttribute('data-tab') === tabId));
    panels.forEach(p => p.classList.toggle('active', p.id === `tab-${tabId}`));
    if (tabId === 'containers') {
      this.refreshContainersStatus();
    }
  }

  async refreshContainersStatus() {
    if (!this.containersList) return;
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const list = await window.go.main.App.GetContainersStatus();
        this.containersList.innerHTML = '';
        list.forEach(c => {
          const ready = c.binary_ready && c.cores_ready;
          const badgeText = ready ? 'CONTAINER READY (ISOLATED)' : 'NOT INITIALIZED';
          const badgeClass = ready ? 'ready' : 'missing';
          const coresInfo = c.installed_cores && c.installed_cores.length > 0
            ? `Cores: ${c.installed_cores.join(', ')} | `
            : '';

          const card = document.createElement('div');
          card.className = 'container-card';
          card.innerHTML = `
            <div class="container-meta">
              <span class="container-name">${c.name}</span>
              <span class="container-systems">${c.systems}</span>
              <span class="container-sub">${coresInfo}Config: ${c.config_path} | ${c.bios_note || ''}</span>
            </div>
            <span class="container-badge ${badgeClass}">${badgeText}</span>
          `;
          this.containersList.appendChild(card);
        });
      } catch (e) {}
    } else {
      // Preview fallback
      this.containersList.innerHTML = `
        <div class="container-card">
          <div class="container-meta">
            <span class="container-name">RetroArch Core Container</span>
            <span class="container-systems">PC Engine CD (Rondo) & PlayStation (SotN)</span>
            <span class="container-sub">Config: emulators/retroarch/config/retroarch.cfg</span>
          </div>
          <span class="container-badge ready">ISOLATED CONTAINER</span>
        </div>
        <div class="container-card">
          <div class="container-meta">
            <span class="container-name">PCSX2 Portable Container</span>
            <span class="container-systems">PlayStation 2 (Lament of Innocence & Curse of Darkness)</span>
            <span class="container-sub">Config: emulators/pcsx2/inis/PCSX2.ini</span>
          </div>
          <span class="container-badge ready">ISOLATED CONTAINER</span>
        </div>
        <div class="container-card">
          <div class="container-meta">
            <span class="container-name">Dolphin Portable Container</span>
            <span class="container-systems">WiiWare (The Adventure ReBirth)</span>
            <span class="container-sub">Config: emulators/dolphin/user/Config/GFX.ini</span>
          </div>
          <span class="container-badge ready">ISOLATED CONTAINER</span>
        </div>
      `;
    }
  }

  async provisionAllContainers() {
    if (!window.go || !window.go.main || !window.go.main.App) return;
    if (this.provisionProgress && this.provisionProgressText) {
      this.provisionProgress.style.display = 'block';
      this.provisionProgressText.textContent = 'Downloading & sandboxing portable emulator containers into emulators/...';
    }
    if (this.provisionAllBtn) {
      this.provisionAllBtn.disabled = true;
      this.provisionAllBtn.textContent = 'INITIALIZING CONTAINERS...';
    }

    try {
      const msg = await window.go.main.App.ProvisionContainers();
      if (this.provisionProgressText) {
        this.provisionProgressText.textContent = msg;
      }
      await this.refreshContainersStatus();
      await this.updateActiveGame();
    } catch (err) {
      if (this.provisionProgressText) {
        this.provisionProgressText.textContent = 'Container setup note: ' + (err?.message || String(err));
      }
    } finally {
      if (this.provisionAllBtn) {
        this.provisionAllBtn.disabled = false;
        this.provisionAllBtn.textContent = 'INITIALIZE / DOWNLOAD CONTAINERS';
      }
    }
  }

  async scanRomFolder() {
    if (!window.go || !window.go.main || !window.go.main.App) return;
    try {
      const res = await window.go.main.App.ScanRomDirectory(true);
      if (this.provisionProgress && this.provisionProgressText) {
        this.provisionProgress.style.display = 'block';
        const biosCount = res.imported_bios ? res.imported_bios.length : 0;
        this.provisionProgressText.textContent = `Scanned ${res.scanned_dir}: Matched ${res.total_matched} ROM(s), synced ${biosCount} BIOS file(s).`;
      }
      await this.refreshContainersStatus();
      await this.updateActiveGame();
    } catch (err) {
      console.error('ROM scan error:', err);
    }
  }

  checkCollectionLogo() {
    const img = new Image();
    img.onload = () => {
      if (this.collectionLogo) this.collectionLogo.style.display = 'block';
      if (this.collectionText) this.collectionText.style.display = 'none';
      if (this.startLogo) this.startLogo.style.display = 'block';
      if (this.startLogoFallback) this.startLogoFallback.style.display = 'none';
    };
    img.onerror = () => {
      if (this.collectionLogo) this.collectionLogo.style.display = 'none';
      if (this.collectionText) this.collectionText.style.display = 'flex';
      if (this.startLogo) this.startLogo.style.display = 'none';
      if (this.startLogoFallback) this.startLogoFallback.style.display = 'flex';
    };
    img.src = 'assets/images/collection_logo.png';
  }

  checkTitleBackground() {
    const img = new Image();
    img.onload = () => {
      if (this.startBg) this.startBg.style.backgroundImage = `url('assets/images/title_background.png')`;
    };
    img.onerror = () => {
      if (this.startBg) this.startBg.style.backgroundImage = `url('assets/images/background.png')`;
    };
    img.src = 'assets/images/title_background.png';
  }

  enterMainMenu() {
    if (!this.inStartScreen) return;
    this.inStartScreen = false;
    if (this.startScreen) {
      this.startScreen.classList.add('hidden');
    }
    this.playSfx('confirm');
    this.startBgm();
  }

  showStartScreen() {
    this.inStartScreen = true;
    if (this.startScreen) {
      this.startScreen.classList.remove('hidden');
    }
  }

  startBgm() {
    if (!this.audioEnabled || !this.bgmPlayer) return;
    if (!this.bgmPlayer.src || !this.bgmPlayer.src.includes('menu_theme.mp3')) {
      this.bgmPlayer.src = 'assets/audio/menu_theme.mp3';
    }
    this.bgmPlayer.loop = true;
    this.bgmPlayer.volume = Math.max(0, Math.min(1, this.menuVolume / 100));
    this.bgmPlayer.play().catch(err => {
      console.log('Audio playback waiting for gesture:', err);
    });
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
    if (!this.uiSfxEnabled || !this.audioCtx) return;
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
      gain.gain.setValueAtTime(0.2, now);
      gain.gain.exponentialRampToValueAtTime(0.01, now + 0.35);
      osc.start(now);
      osc.stop(now + 0.35);
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

  tryLoadFirst(candidates, onMatch, onFail) {
    let i = 0;
    const next = () => {
      if (i >= candidates.length) {
        if (onFail) onFail();
        return;
      }
      const candidate = candidates[i++];
      const test = new Image();
      test.onload = () => onMatch(candidate);
      test.onerror = () => next();
      test.src = candidate;
    };
    next();
  }

  loadCustomAssets(game, region) {
    const activeGameId = game.id;
    const reg = region.toLowerCase();

    // 1. Check Custom Game Title Logo (supports .png, .webp, .jpg, .jpeg)
    if (this.gameLogoImg && this.gameTitle) {
      this.gameLogoImg.style.display = 'none';
      this.gameTitle.style.display = 'block';

      const logoCandidates = [
        `assets/games/${game.id}/logo.png`,
        `assets/games/${game.id}/logo.webp`,
        `assets/games/${game.id}/logo.jpg`,
        `assets/games/${game.id}/logo.jpeg`,
        `assets/games/${game.id}/title.png`
      ];

      this.tryLoadFirst(logoCandidates, (matchedPath) => {
        if (this.games[this.currentIndex]?.id === activeGameId) {
          this.gameLogoImg.src = matchedPath;
          this.gameLogoImg.setAttribute('data-game', game.id);
          this.gameLogoImg.style.display = 'block';
          this.gameTitle.style.display = 'none';
        }
      }, () => {
        if (this.games[this.currentIndex]?.id === activeGameId) {
          this.gameLogoImg.style.display = 'none';
          this.gameTitle.style.display = 'block';
        }
      });
    }

    // 2. Check Box Art (supports box_us, box_en, box, cover in .png, .jpg, .jpeg, .webp)
    const boxCandidates = [
      `assets/games/${game.id}/box_${reg}.png`,
      `assets/games/${game.id}/box_${reg}.jpg`,
      `assets/games/${game.id}/box_${reg}.jpeg`,
      `assets/games/${game.id}/box_${reg}.webp`,
      `assets/games/${game.id}/box_us.png`,
      `assets/games/${game.id}/box_us.jpg`,
      `assets/games/${game.id}/box_us.jpeg`,
      `assets/games/${game.id}/box_us.webp`,
      `assets/games/${game.id}/box_en.png`,
      `assets/games/${game.id}/box_en.jpg`,
      `assets/games/${game.id}/box_en.jpeg`,
      `assets/games/${game.id}/box_en.webp`,
      `assets/games/${game.id}/box.png`,
      `assets/games/${game.id}/box.jpg`,
      `assets/games/${game.id}/cover.png`,
      `assets/games/${game.id}/cover.jpg`
    ];

    this.tryLoadFirst(boxCandidates, (matchedPath) => {
      if (this.games[this.currentIndex]?.id === activeGameId) {
        this.boxartImg.src = matchedPath;
        this.boxartImg.style.display = 'block';
        this.boxartFallback.style.display = 'none';
      }
    }, () => {
      if (this.games[this.currentIndex]?.id === activeGameId) {
        this.boxartImg.style.display = 'none';
        this.boxartFallback.style.display = 'flex';
      }
    });

    // 3. Check Game Background
    const bgCandidates = [
      `assets/games/${game.id}/background.png`,
      `assets/games/${game.id}/background.jpg`,
      `assets/games/${game.id}/bg.png`,
      `assets/games/${game.id}/bg.jpg`,
      `assets/images/background.png`,
      `assets/images/background.jpg`
    ];

    this.tryLoadFirst(bgCandidates, (matchedPath) => {
      if (this.games[this.currentIndex]?.id === activeGameId) {
        this.bgLayer.style.backgroundImage = `url('${matchedPath}')`;
      }
    }, () => {
      if (this.games[this.currentIndex]?.id === activeGameId) {
        this.bgLayer.style.backgroundImage = `url('assets/images/background.png')`;
      }
    });
  }

  showDialog(title, message, tip, showContainerAction = false) {
    if (!this.dialogOverlay) return;
    this.dialogTitle.textContent = title;
    this.dialogMessage.textContent = message;
    if (tip) {
      this.dialogTip.textContent = tip;
      this.dialogTip.style.display = 'block';
    } else {
      this.dialogTip.style.display = 'none';
    }
    if (this.dialogActionBtn) {
      this.dialogActionBtn.style.display = showContainerAction ? 'inline-block' : 'none';
    }
    this.dialogOverlay.style.display = 'flex';
  }

  hideDialog() {
    if (this.dialogOverlay) {
      this.dialogOverlay.style.display = 'none';
    }
  }

  async checkStatus(game, region) {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const status = await window.go.main.App.CheckGameStatus(game.id, region);
        if (status.available && status.runner_ready) {
          this.statusIndicator.querySelector('.status-dot').className = 'status-dot ready';
          this.statusText.textContent = 'READY TO PLAY';
          this.missingHint.style.display = 'none';
        } else if (status.resolved_rom && !status.runner_ready) {
          this.statusIndicator.querySelector('.status-dot').className = 'status-dot missing';
          this.statusText.textContent = 'CONTAINER SETUP NEEDED';
          this.missingHint.style.display = 'flex';
          this.missingHintText.textContent = status.missing_runner || 'Open SETTINGS -> CONTAINERS & ROMS to initialize isolated runner';
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
        console.error('Launch failed:', err);
        const errMsg = err?.message || String(err) || 'Error launching game.';
        const isContainerMissing = errMsg.includes('container') || errMsg.includes('core') || errMsg.includes('INITIALIZE');
        const tip = isContainerMissing
          ? 'Click "OPEN CONTAINER SETUP" below to automatically download and sandbox the portable emulators inside emulators/ without affecting your system.'
          : 'Verify that your ROM file is placed inside games/' + game.id + '/';
        this.showDialog('LAUNCH STATUS', errMsg, tip, isContainerMissing);
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
    this.settings.menu_music = this.audioEnabled;
    this.audioLabel.textContent = this.audioEnabled ? 'AUDIO ON' : 'AUDIO OFF';
    if (!this.audioEnabled) {
      this.bgmPlayer.pause();
    } else {
      this.startBgm();
    }
  }

  bindEvents() {
    if (this.startScreen) {
      this.startScreen.addEventListener('click', () => this.enterMainMenu());
    }

    if (this.dialogCloseBtn) {
      this.dialogCloseBtn.addEventListener('click', () => this.hideDialog());
    }
    if (this.dialogActionBtn) {
      this.dialogActionBtn.addEventListener('click', () => {
        this.hideDialog();
        this.openSettings('containers');
      });
    }
    if (this.dialogOverlay) {
      this.dialogOverlay.addEventListener('click', (e) => {
        if (e.target === this.dialogOverlay) this.hideDialog();
      });
    }

    // Settings Modal Events
    if (this.settingsBtn) {
      this.settingsBtn.addEventListener('click', () => this.openSettings());
    }
    if (this.settingsCloseBtn) {
      this.settingsCloseBtn.addEventListener('click', () => this.closeSettings());
    }
    if (this.saveSettingsBtn) {
      this.saveSettingsBtn.addEventListener('click', () => this.saveSettings(true));
    }
    if (this.settingsOverlay) {
      this.settingsOverlay.addEventListener('click', (e) => {
        if (e.target === this.settingsOverlay) this.closeSettings();
      });
    }

    document.querySelectorAll('.settings-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        this.playSfx('cursor');
        this.switchSettingsTab(tab.getAttribute('data-tab'));
      });
    });

    if (this.optMenuVolume) {
      this.optMenuVolume.addEventListener('input', (e) => {
        const val = parseInt(e.target.value, 10);
        this.menuVolume = val;
        if (this.volValueLabel) this.volValueLabel.textContent = `${val}%`;
        if (this.bgmPlayer) this.bgmPlayer.volume = Math.max(0, Math.min(1, val / 100));
      });
    }

    // Auto-apply setting changes immediately
    [
      this.optDisplayMode, this.optResolution, this.optAspectRatio,
      this.optRenderer, this.optShaderFilter, this.optVsync,
      this.optShowFps, this.optMenuMusic, this.optUiSfx, this.optAudioSync
    ].forEach(el => {
      if (el) {
        el.addEventListener('change', () => this.saveSettings(false));
      }
    });

    if (this.provisionAllBtn) {
      this.provisionAllBtn.addEventListener('click', () => this.provisionAllContainers());
    }
    if (this.scanRomsBtn) {
      this.scanRomsBtn.addEventListener('click', () => this.scanRomFolder());
    }

    this.launchBtn.addEventListener('click', () => this.launchCurrentGame());
    this.fullscreenBtn.addEventListener('click', () => this.toggleFullscreen());
    this.audioToggleBtn.addEventListener('click', () => this.toggleAudio());

    window.addEventListener('keydown', (e) => {
      if (this.inStartScreen) {
        this.enterMainMenu();
        return;
      }

      if (this.inSettings) {
        if (e.key === 'Escape') {
          this.closeSettings();
        }
        return;
      }

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
        case 's':
        case 'S':
        case 'o':
        case 'O':
          this.openSettings();
          break;
        case 'Escape':
          this.showStartScreen();
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
        const anyButtonPressed = gp.buttons.some(b => b && b.pressed);

        if (this.inStartScreen) {
          if (anyButtonPressed && now - this.gamepadTimestamp > 300) {
            this.enterMainMenu();
            this.gamepadTimestamp = now + 400;
          }
          requestAnimationFrame(poll);
          return;
        }

        if (now - this.gamepadTimestamp > 220) {
          const axisX = gp.axes[0];
          const dpadLeft = gp.buttons[14] && gp.buttons[14].pressed;
          const dpadRight = gp.buttons[15] && gp.buttons[15].pressed;
          const dpadUp = gp.buttons[12] && gp.buttons[12].pressed;
          const dpadDown = gp.buttons[13] && gp.buttons[13].pressed;

          if (this.inSettings) {
            if (gp.buttons[1] && gp.buttons[1].pressed) {
              this.closeSettings();
              this.gamepadTimestamp = now + 350;
            }
            requestAnimationFrame(poll);
            return;
          }

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
          } else if (gp.buttons[1] && gp.buttons[1].pressed) {
            this.showStartScreen();
            this.gamepadTimestamp = now + 400;
          } else if (gp.buttons[8] && gp.buttons[8].pressed) {
            this.openSettings();
            this.gamepadTimestamp = now + 400;
          } else if (gp.buttons[9] && gp.buttons[9].pressed) {
            this.openSettings();
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
