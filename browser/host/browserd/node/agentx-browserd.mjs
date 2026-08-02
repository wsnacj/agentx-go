#!/usr/bin/env node

import { createHash, randomUUID } from 'node:crypto';
import { existsSync, readFileSync } from 'node:fs';
import fs from 'node:fs/promises';
import http from 'node:http';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const backendName = 'agentx-browserd';
const attachCDPEndpoint = envString('AGENTX_BROWSERD_ATTACH_CDP_ENDPOINT');
const attachCDPMode = attachCDPEndpoint !== '';
const browserApp = attachCDPMode ? 'Chrome' : 'Chromium';
const defaultProfile = 'default';
const scriptRoot = path.dirname(fileURLToPath(import.meta.url));
const cssElementRefPrefix = 'css1:';
const metaElementRefPrefix = 'elem1:';
const mockMode = envString('AGENTX_BROWSERD_TEST_MODE') === 'mock';
const enableTestHooks = envString('AGENTX_BROWSERD_ENABLE_TEST_HOOKS') === '1';
const host = envString('AGENTX_BROWSERD_HOST') || '127.0.0.1';
const port = envInt('AGENTX_BROWSERD_PORT');
const token = envString('AGENTX_BROWSERD_TOKEN');
const stateRoot = envString('AGENTX_BROWSERD_STATE_ROOT') || path.resolve('.agentx/browserd');
const profilesRoot = envString('AGENTX_BROWSERD_PROFILES_ROOT') || path.join(stateRoot, 'profiles');
const artifactsRoot = envString('AGENTX_BROWSERD_ARTIFACTS_ROOT') || path.join(stateRoot, 'artifacts');
const logsRoot = envString('AGENTX_BROWSERD_LOGS_ROOT') || path.join(stateRoot, 'logs');
const playwrightBrowsersPath = envString('PLAYWRIGHT_BROWSERS_PATH');
const playwrightCacheSource = envString('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_SOURCE');
const playwrightCachePinned = envString('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_PINNED') === '1';
const playwrightBundleGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_BUNDLE_GENERATION');
const playwrightDependencyGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_DEPENDENCY_GENERATION');
const playwrightBrowserRevision = envString('AGENTX_BROWSERD_PLAYWRIGHT_BROWSER_REVISION');
const playwrightDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_GENERATION');
const playwrightTargetDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_TARGET_DELIVERY_GENERATION');
const playwrightLastReadyDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_LAST_READY_DELIVERY_GENERATION');
const playwrightRetainedDeliveries = splitCSV(envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERIES'));
const playwrightLastEvictedDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_LAST_EVICTED_DELIVERY_GENERATION');
const playwrightLastDeliverySwitchUnix = envInt('AGENTX_BROWSERD_PLAYWRIGHT_LAST_DELIVERY_SWITCH_UNIX_MILLI');
const playwrightRetainedDeliveryRevision = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERY_BROWSER_REVISION');
const playwrightRetainedDeliveryReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_DELIVERY_CACHE_READY') === '1';
const playwrightRetainedFallbackDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_DELIVERY_GENERATION');
const playwrightRetainedFallbackPayloadReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_READY') === '1';
const playwrightRetainedFallbackPayloadBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_BLOCK_REASON');
const playwrightRetainedFallbackPayloadSource = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_SOURCE');
const playwrightRetainedFallbackPayloadDirs = splitCSV(envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_PAYLOAD_DIRS'));
const playwrightRetainedFallbackLaunchReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_LAUNCH_READY') === '1';
const playwrightRetainedFallbackLaunchBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_RETAINED_FALLBACK_LAUNCH_BLOCK_REASON');
const playwrightSelectedLaunchDeliveryGeneration = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_DELIVERY_GENERATION');
const playwrightSelectedLaunchSource = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_SOURCE');
const playwrightSelectedLaunchReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_READY') === '1';
const playwrightSelectedLaunchBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_BLOCK_REASON');
const playwrightSelectedLaunchBrowserRevision = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_BROWSER_REVISION');
const playwrightSelectedLaunchPayloadSource = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_SOURCE');
const playwrightSelectedLaunchPayloadDirs = splitCSV(envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_DIRS'));
const playwrightSelectedLaunchPayloadReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_READY') === '1';
const playwrightSelectedLaunchPayloadBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_PAYLOAD_BLOCK_REASON');
const playwrightSelectedLaunchExecutablePath = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_PATH');
const playwrightSelectedLaunchExecutableReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_READY') === '1';
const playwrightSelectedLaunchExecutableBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_SELECTED_LAUNCH_EXECUTABLE_BLOCK_REASON');
const playwrightDeliveryTransitionPending = envString('AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_TRANSITION_PENDING') === '1';
const playwrightDeliveryTransitionStage = envString('AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_TRANSITION_STAGE');
const playwrightLaunchReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_LAUNCH_READY') === '1';
const playwrightLaunchBlockReason = envString('AGENTX_BROWSERD_PLAYWRIGHT_LAUNCH_BLOCK_REASON');
const playwrightBundleReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_BUNDLE_READY') === '1';
const playwrightDeliveryReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_DELIVERY_READY') === '1';
const playwrightCachePolicyVersion = envString('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_POLICY_VERSION');
const playwrightCacheRetentionMode = envString('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_RETENTION_MODE');
const playwrightCacheRetainedDirs = splitCSV(envString('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_RETAINED_DIRS'));
const playwrightCacheLastGCPrunedDirCount = envInt('AGENTX_BROWSERD_PLAYWRIGHT_CACHE_LAST_GC_PRUNED_DIR_COUNT');
const playwrightBootstrapState = envString('AGENTX_BROWSERD_PLAYWRIGHT_BOOTSTRAP_STATE');
const playwrightBootstrapErrorCode = envString('AGENTX_BROWSERD_PLAYWRIGHT_BOOTSTRAP_ERROR_CODE');
const playwrightNodeModulesReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_NODE_MODULES_READY') === '1';
const playwrightBrowserReady = envString('AGENTX_BROWSERD_PLAYWRIGHT_BROWSER_READY') === '1';
const sessionEventBufferLimit = 32;
const disconnectBurstWindowMs = 120000;
const disconnectBurstThreshold = 2;
const restartFailurePermanentThreshold = 2;
const recentDownloadReuseWindowMs = 60000;

if (!port) {
  console.error('AGENTX_BROWSERD_PORT is required');
  process.exit(1);
}

const runtimeState = {
  browser: null,
  context: null,
  activePage: null,
  activeProfile: defaultProfile,
  mockRunning: true,
  playwright: null,
  playwrightPackage: '',
  profiles: new Map([[defaultProfile, { profile: defaultProfile }]]),
  pageNativeRefs: new Map(),
  pageSessionState: new Map(),
  retainedSessionStateByProfile: new Map(),
  browserCloseIntent: false,
  nextNativeRefID: 1,
  managedPages: new WeakSet()
};
const pageTargetIDs = new WeakMap();

const server = http.createServer(async (req, res) => {
  try {
    if (req.method !== 'POST') {
      sendError(res, 405, 'method not allowed');
      return;
    }
    if (!authorized(req)) {
      sendError(res, 401, 'unauthorized');
      return;
    }
    const payload = await readJSON(req);
    const method = stringValue(payload?.method);
    const params = normalizeRequestParams(payload?.params);
    if (!method) {
      sendError(res, 400, 'missing method');
      return;
    }
    const result = await handleMethod(method, params);
    sendResult(res, result);
  } catch (err) {
    sendError(res, errorStatus(err), errorMessage(err), errorResolverOutcome(err));
  }
});

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

await ensureRoots();
server.listen(port, host, () => {
  console.log(`agentx-browserd listening on http://${host}:${port}`);
});

async function shutdown() {
  server.close();
  await closeBrowser();
  process.exit(0);
}

async function handleMethod(method, params) {
  switch (method) {
    case 'browser.status':
      return buildStatusResult(stringValue(params.profile));
    case 'browser.shutdown':
      return requestShutdown();
    case 'browser.start':
      return startProfile(stringValue(params.profile));
    case 'browser.stop':
      return stopProfile(stringValue(params.profile));
    case 'browser.profiles':
      return listProfiles();
    case 'browser.profile.create':
      return createProfile(params);
    case 'browser.profile.delete':
      return deleteProfile(params);
    case 'browser.open':
      return openPage(params);
    case 'browser.navigate':
      return navigatePage(params);
    case 'browser.tabs':
      return handleTabs(params);
    case 'browser.extract':
      return extractPage(params);
    case 'browser.snapshot':
      return snapshotPage(params);
    case 'browser.screenshot':
      return screenshotPage(params);
    case 'browser.errors':
      return pageErrors(params);
    case 'browser.download':
      return downloadPage(params);
    case 'browser.wait_download':
      return waitDownload(params);
    case 'browser.dialog':
      return armDialog(params);
    case 'browser.highlight':
      return highlightPage(params);
    case 'browser.upload':
      return uploadPage(params);
    case 'browser.fill':
      return fillPage(params);
    case 'browser.select':
      return selectOptionPage(params);
    case 'browser.hover':
      return hoverPage(params);
    case 'browser.drag':
      return dragPage(params);
    case 'browser.click':
      return clickPage(params);
    case 'browser.type':
      return typeIntoPage(params);
    case 'browser.eval':
      return evalPage(params);
    case 'browser.artifact.resolve':
      return resolveArtifact(params);
    case 'browser.test.action_resolver':
      return inspectActionResolver(params);
    default:
      throw httpError(400, `unsupported method: ${method}`);
  }
}

async function ensureRoots() {
  for (const dir of [stateRoot, profilesRoot, artifactsRoot, logsRoot]) {
    await fs.mkdir(dir, { recursive: true });
  }
  await ensureProfileDir(defaultProfile);
}

function buildStatusResult(profileName) {
  const profile = ensureProfile(profileName);
  const running = mockMode ? (runtimeState.mockRunning && normalizedProfile(runtimeState.activeProfile) === profile) : isProfileRunning(profile);
  const connected = mockMode ? (runtimeState.mockRunning && normalizedProfile(runtimeState.activeProfile) === profile) : isProfileRunning(profile);
  const page = currentPageForProfile(profile);
  const sessionHealth = sessionHealthSummaryForState(currentSessionStateForProfile(profile), page);
  let runtimeStatus = running ? 'running' : 'idle';
  if (!running && !connected && ['browser_disconnected', 'browser_disconnect_burst', 'cooldown_active', 'restart_pending', 'restart_failed', 'restart_failed_permanent'].includes(stringValue(sessionHealth?.state))) {
    runtimeStatus = 'disconnected';
  }
  const noteParts = browserStatusNoteTokens(profile);
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: profile,
    state_root: stateRoot,
    profiles_root: profilesRoot,
    artifacts_root: artifactsRoot,
    logs_root: logsRoot,
    status: runtimeStatus,
    running: running,
    connected: connected,
    playwright_cache: playwrightCacheSummary(),
    session_health: sessionHealth,
    note: noteParts.join(' ')
  };
}

function requestShutdown() {
  setTimeout(() => {
    shutdown().catch((err) => {
      console.error(err);
      process.exit(1);
    });
  }, 0);
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: normalizedProfile(runtimeState.activeProfile) || defaultProfile,
    state_root: stateRoot,
    profiles_root: profilesRoot,
    artifacts_root: artifactsRoot,
    logs_root: logsRoot,
    status: 'stopping',
    running: !!runtimeState.browser,
    connected: !!runtimeState.context,
    note: 'shutdown_requested'
  };
}

async function startProfile(profileName) {
  const profile = normalizedProfile(profileName);
  enforceAttachProfile(profile);
  ensureProfile(profile);
  if (mockMode) {
    runtimeState.mockRunning = true;
    runtimeState.activeProfile = profile;
  } else {
    await ensureProfileDir(profile);
    await ensureBrowser(profile, { allowEscalatedRestart: true });
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: profile,
    status: 'running',
    running: true,
    connected: true,
    note: mockMode ? 'mock started' : 'started'
  };
}

async function stopProfile(profileName) {
  const profile = normalizedProfile(profileName);
  enforceAttachProfile(profile);
  ensureProfile(profile);
  if (mockMode) {
    if (normalizedProfile(runtimeState.activeProfile) === profile) {
      runtimeState.mockRunning = false;
    }
  } else if (isProfileRunning(profile)) {
    await persistActiveProfileState();
    await closeActiveProfileContext({ closeBrowser: !attachCDPMode });
    if (!attachCDPMode) {
      runtimeState.browser = null;
    }
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: profile,
    status: 'stopped',
    running: false,
    connected: false,
    note: 'stopped'
  };
}

function listProfiles() {
  const activeProfile = normalizedProfile(runtimeState.activeProfile);
  const profiles = attachCDPMode ? [defaultProfile] : Array.from(runtimeState.profiles.keys()).sort();
  return {
    backend: backendName,
    default_profile: defaultProfile,
    profiles: profiles.map((profile) => ({
      profile: profile,
      browser_app: browserApp,
      status: (mockMode ? (runtimeState.mockRunning && activeProfile === profile) : isProfileRunning(profile)) ? 'running' : 'idle',
      running: mockMode ? (runtimeState.mockRunning && activeProfile === profile) : isProfileRunning(profile),
      connected: mockMode ? (runtimeState.mockRunning && activeProfile === profile) : isProfileRunning(profile),
      note: profile === activeProfile
        ? (mockMode ? 'mock active profile' : (isProfileRunning(profile) ? 'active profile' : 'selected profile'))
        : (profile === defaultProfile ? '' : 'created')
    }))
  };
}

async function createProfile(params) {
  const current = normalizeRequestParams(params);
  const profile = normalizedProfile(stringValue(current.profile));
  const copyFrom = stringValue(current.copy_from);
  if (attachCDPMode && profile !== defaultProfile) {
    throw httpError(400, 'host CDP attach only supports the default profile');
  }
  ensureProfile(profile);
  await ensureProfileDir(profile);
  if (copyFrom && normalizedProfile(copyFrom) !== profile) {
    await copyProfileStorageState(copyFrom, profile);
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: profile,
    status: 'created',
    running: false,
    connected: false,
    note: copyFrom && normalizedProfile(copyFrom) !== profile ? `created from ${normalizedProfile(copyFrom)}` : 'created'
  };
}

async function deleteProfile(params) {
  const current = normalizeRequestParams(params);
  const profile = normalizedProfile(stringValue(current.profile));
  const force = Boolean(current.force);
  if (attachCDPMode) {
    throw httpError(400, 'host CDP attach does not support profile deletion');
  }
  ensureProfile(profile);
  if (profile === defaultProfile) {
    throw httpError(400, 'default profile cannot be deleted');
  }
  if (isProfileRunning(profile) && !force) {
    throw httpError(409, `profile is running: ${profile}`);
  }
  if (isProfileRunning(profile)) {
    await stopProfile(profile);
  }
  runtimeState.profiles.delete(profile);
  if (normalizedProfile(runtimeState.activeProfile) === profile) {
    runtimeState.activeProfile = defaultProfile;
  }
  await fs.rm(path.join(profilesRoot, profile), { recursive: true, force: true });
  return {
    backend: backendName,
    browser_app: browserApp,
    profile: profile,
    status: 'deleted',
    running: false,
    connected: false,
    note: 'deleted'
  };
}

async function openPage(params) {
  const page = await preparePage(params, { allowBlank: true, createPage: true });
  clearStaleTargetState(page);
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    Status: 'ok',
    Note: page.url() || 'about:blank'
  };
}

async function navigatePage(params) {
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, { allowBlank: false, createPage: true });
  const url = stringValue(params.url);
  let downloadStarted = false;
  let navigationWaitFallback = '';
  if (url) {
    const navigationResult = await gotoPageWithLoadFallback(page, url, waitTimeout(params));
    downloadStarted = navigationResult.downloadStarted === true;
    navigationWaitFallback = stringValue(navigationResult.waitFallback);
  } else if (page.url() === 'about:blank') {
    throw httpError(400, 'url is required for this action until a page has been opened');
  }
  if (intValue(params.wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.wait_ms));
  }
  runtimeState.activePage = page;
  clearStaleTargetState(page);
  const noteTokens = [];
  if (downloadStarted) {
    noteTokens.push(noteToken('download_started', 'true'));
    if (url) {
      noteTokens.push(noteToken('requested_url', url));
    }
  }
  if (navigationWaitFallback) {
    noteTokens.push(noteToken('navigation_wait_fallback', navigationWaitFallback));
  }
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    FinalURL: downloadStarted && url ? url : page.url(),
    Title: await page.title(),
    Status: 'ok',
    Note: joinNoteTokens(noteTokens)
  };
}

async function handleTabs(params) {
  if (mockMode) {
    return {
      Backend: backendName,
      BrowserApp: browserApp,
      Action: stringValue(params.action) || 'list',
      Status: 'ok',
      Tabs: [{
        Index: 1,
        Title: 'Mock Tab',
        URL: stringValue(params.url) || 'about:blank',
        target_id: 'mock-target-1',
        Active: true
      }],
      ActiveIndex: 1
    };
  }
  await ensureBrowser();
  const pages = currentPages();
  const action = stringValue(params.action) || 'list';
  if (action === 'focus') {
    const page = pageByTabIndex(pages, intValue(params.tab_index));
    runtimeState.activePage = page;
    await page.bringToFront();
  } else if (action === 'close') {
    const page = pageByTabIndex(pages, intValue(params.tab_index));
    await page.close();
    runtimeState.activePage = currentPages()[0] || null;
  }
  const listedPages = currentPages();
  const activePage = runtimeState.activePage || listedPages[0] || null;
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    Action: action,
    Status: 'ok',
      Tabs: await Promise.all(listedPages.map(async (page, idx) => ({
        Index: idx + 1,
        Title: await page.title(),
        URL: page.url(),
        target_id: targetIDForPage(page),
        Active: page === activePage
      }))),
    ActiveIndex: activePage ? listedPages.indexOf(activePage) + 1 : 0
  };
}

async function extractPage(params) {
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const maxChars = intValue(params.max_chars);
  const content = await page.locator('body').innerText();
  clearStaleTargetState(page);
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    Title: await page.title(),
    Content: truncate(content, maxChars),
    FinalURL: page.url(),
    ContentType: 'text/plain'
  };
}

async function snapshotPage(params) {
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const maxChars = intValue(params.max_chars);
  const maxElements = intValue(params.max_elements) || 32;
  const selector = stringValue(params.selector);
  const snapshot = await captureSnapshotForPage(page, {
    requestedSelector: selector,
    limitChars: maxChars,
    limitElements: maxElements
  });
  const finalURL = page.url();
  const title = await page.title();
  const pageBinding = {
    page_url: finalURL,
    page_origin: urlOrigin(finalURL),
    page_path: urlPath(finalURL),
    page_title: title,
    tab_index: currentTabIndexForPage(page)
  };
  const elements = Array.isArray(snapshot.elements) ? snapshot.elements.map((element) => {
    const ref = registerSnapshotNativeRef(page, element, pageBinding);
    if (!ref) {
      return element;
    }
    return { ...element, Ref: ref };
  }) : [];
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: finalURL,
    title: title,
    snapshot: snapshot.snapshot || '',
    elements,
    truncated: Boolean(snapshot.truncated),
    format: 'text',
    mode: 'minimal',
    refs: elements.some((element) => stringValue(element.Ref) !== '') ? 'native_ref' : 'none',
    interactive: true,
    selector: selector,
    note: stringValue(snapshot.note)
  };
}

async function captureSnapshotForPage(page, options) {
  const selector = stringValue(options.requestedSelector);
  const limitChars = intValue(options.limitChars);
  const limitElements = intValue(options.limitElements) > 0 ? intValue(options.limitElements) : 32;
  if (selector) {
    const target = await findSnapshotFrameTarget(page, selector);
    if (!target) {
      return {
        snapshot: '',
        elements: [],
        note: `selector not found: ${selector}`
      };
    }
    const partial = await snapshotFrameContent(target.frame, {
      requestedSelector: selector,
      limitChars,
      limitElements
    });
    return {
      snapshot: stringValue(partial.snapshot),
      truncated: Boolean(partial.truncated),
      elements: annotateSnapshotElementsForFrame(partial.elements, target.path),
      note: stringValue(partial.note)
    };
  }

  const frames = frameTraversalDescriptors(page);
  const elements = [];
  let snapshotText = '';
  let truncated = false;
  let note = '';
  let remainingElements = limitElements;
  for (const descriptor of frames) {
    if (remainingElements <= 0) {
      break;
    }
    const partial = await snapshotFrameContent(descriptor.frame, {
      requestedSelector: '',
      limitChars: descriptor.path === '' ? limitChars : 0,
      limitElements: remainingElements
    });
    if (descriptor.path === '') {
      snapshotText = stringValue(partial.snapshot);
      truncated = Boolean(partial.truncated);
      note = stringValue(partial.note);
    }
    const annotated = annotateSnapshotElementsForFrame(partial.elements, descriptor.path);
    elements.push(...annotated);
    remainingElements = Math.max(0, remainingElements - annotated.length);
  }
  return {
    snapshot: snapshotText,
    truncated,
    elements,
    note
  };
}

async function findSnapshotFrameTarget(page, selector) {
  for (const descriptor of frameTraversalDescriptors(page)) {
    try {
      const found = await descriptor.frame.evaluate(snapshotFrameSelectorExistsEvaluate, {
        requestedSelector: selector
      });
      if (found) {
        return descriptor;
      }
    } catch {
      continue;
    }
  }
  return null;
}

function frameTraversalDescriptors(page) {
  const out = [];
  const mainFrame = page.mainFrame();
  visit(mainFrame, '');
  return out;

  function visit(frame, path) {
    if (!frame) {
      return;
    }
    out.push({ frame, path });
    const children = frame.childFrames();
    for (let idx = 0; idx < children.length; idx += 1) {
      visit(children[idx], path ? `${path}/${idx}` : `${idx}`);
    }
  }
}

async function snapshotFrameContent(frame, payload) {
  try {
    const result = await frame.evaluate(snapshotFrameEvaluate, payload);
    return objectValue(result);
  } catch {
    return {
      snapshot: '',
      elements: [],
      note: ''
    };
  }
}

function annotateSnapshotElementsForFrame(elements, framePath) {
  const normalizedFramePath = stringValue(framePath);
  return arrayValue(elements).map((element) => ({
    ...element,
    FramePath: normalizedFramePath,
    frame_path: normalizedFramePath
  }));
}

async function screenshotPage(params) {
  if (mockMode) {
    const targetPath = await writeMockScreenshot();
    return {
      backend: backendName,
      browser_app: browserApp,
      path: targetPath,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Screenshot',
      capture_scope: captureScope(params),
      capture_width: 1,
      capture_height: 1,
      status: 'ok',
      note: 'mock'
    };
  }
  const page = await preparePage(params, { allowBlank: true, createPage: true });
  const target = await resolveActionLocatorForPage(page, params, { required: actionTargetRequested(params) });
  const actionability = target.locator
    ? await browserActionabilityReportForLocator('screenshot', target.locator, target.outcome, params)
    : null;
  const targetPath = artifactPath('screenshot', 'png');
  let screenshotBuffer;
  try {
    if (target.locator) {
      screenshotBuffer = await target.locator.screenshot({ path: targetPath });
    } else {
      screenshotBuffer = await page.screenshot({
        path: targetPath,
        fullPage: Boolean(params.full_page)
      });
    }
  } catch (err) {
    return browserActionFailureResult('screenshot', page, target, actionability, err, {
      path: targetPath
    });
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  const size = pngSizeFromBuffer(screenshotBuffer);
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    path: targetPath,
    final_url: page.url(),
    title: await page.title(),
    capture_scope: captureScope(params),
    capture_width: size.width,
    capture_height: size.height,
    status: 'ok',
    actionability,
    resolver_outcome: target.outcome
  };
}

async function highlightPage(params) {
  const ref = firstNonEmpty(params.ref, params.element_ref);
  const selector = stringValue(params.selector);
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      ref,
      selector,
      status: 'highlighted',
      note: 'mock highlight'
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const target = await resolveActionLocatorForPage(page, params, { required: true });
  const actionability = await browserActionabilityReportForLocator('highlight', target.locator, target.outcome, params);
  try {
    await target.locator.evaluate((element) => {
      if (!(element instanceof Element)) {
        throw new Error('target is not an element');
      }
      element.scrollIntoView({ block: 'center', inline: 'center', behavior: 'auto' });
      element.setAttribute('data-agentx-highlight', 'true');
      const existingOutline = element.style.outline;
      const existingOutlineOffset = element.style.outlineOffset;
      const existingBoxShadow = element.style.boxShadow;
      element.setAttribute('data-agentx-highlight-prev-outline', existingOutline || '');
      element.setAttribute('data-agentx-highlight-prev-outline-offset', existingOutlineOffset || '');
      element.setAttribute('data-agentx-highlight-prev-box-shadow', existingBoxShadow || '');
      element.style.outline = '3px solid #f59e0b';
      element.style.outlineOffset = '2px';
      element.style.boxShadow = '0 0 0 4px rgba(245, 158, 11, 0.25)';
    });
  } catch (err) {
    return browserActionFailureResult('highlight', page, target, actionability, err, {
      ref,
      selector
    });
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    ref,
    selector,
    status: 'highlighted',
    actionability,
    resolver_outcome: target.outcome
  };
}

async function pageErrors(params) {
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      errors: [],
      session_health: null,
      status: params.clear ? 'cleared' : 'ok',
      note: 'mock'
    };
  }
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, { createPage: false });
  if (intValue(params.wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.wait_ms));
  }
  const state = pageSessionStateForPage(page);
  const entries = browserErrorEntriesForState(state);
  if (Boolean(params.clear)) {
    state.events = [];
  }
  runtimeState.activePage = page;
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    errors: entries,
    session_health: sessionHealthSummaryForState(state, page),
    status: Boolean(params.clear) ? 'cleared' : 'ok',
    note: sessionDiagnosticsNote(page)
  };
}

async function downloadPage(params) {
  if (mockMode) {
    const targetPath = await writeMockDownload(stringValue(params.output_path));
    return {
      Backend: backendName,
      BrowserApp: browserApp,
      Path: targetPath,
      FinalURL: stringValue(params.url) || 'about:blank',
      Title: 'Mock Download',
      ContentType: contentTypeFromHints(targetPath),
      Note: 'mock download'
    };
  }
  const url = stringValue(params.url);
  if (!url) {
    throw httpError(400, 'url is required');
  }
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, { createPage: true });
  const timeout = waitTimeout(params);
  const baseline = pageSessionStateForPage(page).nextDownloadSeq;
  const nextDownload = waitForDownloadEntry(page, baseline, timeout);
  void page.goto(url, { waitUntil: 'commit', timeout }).catch(() => {});
  const downloadWait = await nextDownload;
  const entry = downloadWait && downloadWait.entry ? downloadWait.entry : downloadWait;
  const record = await finalizeDownloadEntry(entry, stringValue(params.output_path));
  const state = pageSessionStateForPage(page);
  state.lastDownload = record;
  state.updatedAt = Date.now();
  runtimeState.activePage = page;
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    Path: record.path,
    FinalURL: record.finalURL,
    Title: record.title,
    ContentType: record.contentType,
    Download: downloadResultMetadata('direct', record),
    Note: joinNoteTokens(downloadResultNoteTokens('direct', record))
  };
}

async function waitDownload(params) {
  if (mockMode) {
    const targetPath = await writeMockDownload(stringValue(params.output_path));
    return {
      Backend: backendName,
      BrowserApp: browserApp,
      Path: targetPath,
      FinalURL: 'about:blank',
      Title: 'Mock Download',
      ContentType: contentTypeFromHints(targetPath),
      Note: 'mock waited for download'
    };
  }
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, { createPage: false });
  const downloadWait = await waitForDownloadEntry(page, 0, waitTimeout(params), {
    allowRecentDownloadReuse: params.allow_recent_download_reuse === true
  });
  const entry = downloadWait && downloadWait.entry ? downloadWait.entry : downloadWait;
  const record = await finalizeDownloadEntry(entry, stringValue(params.output_path));
  const state = pageSessionStateForPage(page);
  const waitMode = downloadWait && downloadWait.mode ? downloadWait.mode : 'wait';
  state.lastDownload = markDownloadRecordWaitConsumed(record, waitMode);
  state.updatedAt = Date.now();
  runtimeState.activePage = page;
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    Path: record.path,
    FinalURL: record.finalURL,
    Title: record.title,
    ContentType: record.contentType,
    Download: downloadResultMetadata(waitMode, record),
    Note: joinNoteTokens(downloadResultNoteTokens(waitMode, record))
  };
}

async function armDialog(params) {
  if (mockMode) {
    return {
      Backend: backendName,
      BrowserApp: browserApp,
      FinalURL: 'about:blank',
      Title: 'Mock Page',
      Status: 'armed',
      Note: 'mock dialog armed'
    };
  }
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, { createPage: false });
  const action = normalizeDialogAction(stringValue(params.action));
  if (!action) {
    throw httpError(400, 'dialog action must be accept or dismiss');
  }
  const session = pageSessionStateForPage(page);
  session.pendingDialog = {
    action,
    promptText: stringValue(params.prompt_text),
    armedAt: Date.now()
  };
  runtimeState.activePage = page;
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    FinalURL: page.url(),
    Title: await page.title(),
    Status: 'armed',
    Dialog: dialogResultMetadata(action, session.pendingDialog.promptText),
    Note: joinNoteTokens(dialogResultNoteTokens(action, session.pendingDialog.promptText))
  };
}

async function uploadPage(params) {
  const paths = uploadPaths(params);
  const note = joinNoteTokens(uploadResultNoteTokens(uploadTargetRequested(params) ? 'input_files' : 'armed', paths.length));
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      status: uploadTargetRequested(params) ? 'uploaded' : 'armed',
      note: firstNonEmpty(note, 'mock upload')
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  runtimeState.activePage = page;
  if (!uploadTargetRequested(params)) {
    const state = pageSessionStateForPage(page);
    state.pendingUpload = {
      paths: [...paths],
      armedAt: Date.now()
    };
    state.updatedAt = Date.now();
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: page.url(),
      title: await page.title(),
      status: 'armed',
      note: note
    };
  }
  const target = await resolveActionLocatorForPage(page, uploadActionLocatorParams(params), { required: true, allowHidden: true });
  const actionability = await browserActionabilityReportForLocator('upload', target.locator, target.outcome, params);
  try {
    await target.locator.setInputFiles(paths);
  } catch (err) {
    return browserActionFailureResult('upload', page, target, actionability, err, { note });
  }
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    status: 'uploaded',
    note: note,
    actionability,
    resolver_outcome: target.outcome
  };
}

async function fillPage(params) {
  const fields = normalizeFillFields(params.fields);
  if (fields.length === 0) {
    throw httpError(400, 'fields is required');
  }
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      field_count: fields.length,
      status: 'filled',
      submitted: Boolean(params.submit),
      note: 'mock fill'
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  let lastOutcome = null;
  let lastLocator = null;
  let lastActionability = null;
  let fallbackUsed = false;
  const timeout = waitTimeout(params);
  for (const field of fields) {
    const target = await resolveActionLocatorForPage(page, fillFieldActionLocatorParams(field), { required: true });
    const fieldActionability = await browserActionabilityReportForLocator('fill', target.locator, target.outcome, field);
    if (actionabilityFailedOnAny(fieldActionability, ['enabled', 'editable'])) {
      return browserActionFailureResult('fill', page, target, fieldActionability, new Error(fieldActionabilityFailureMessage(fieldActionability)), {
        field_count: fields.length
      });
    }
    let fieldResult;
    try {
      fieldResult = await applyFillField(target.locator, field, timeout);
    } catch (err) {
      return browserActionFailureResult('fill', page, target, fieldActionability, err, {
        field_count: fields.length
      });
    }
    fallbackUsed = fallbackUsed || Boolean(fieldResult?.fallbackUsed);
    lastOutcome = target.outcome || lastOutcome;
    lastLocator = target.locator;
    lastActionability = fieldActionability || lastActionability;
  }
  let submitted = false;
  let submitFallbackUsed = false;
  if (Boolean(params.submit) && lastLocator) {
    try {
      const submitResult = await submitLocatorWithFallback(lastLocator, timeout);
      submitted = true;
      submitFallbackUsed = Boolean(submitResult?.fallbackUsed);
    } catch (err) {
      return browserActionFailureResult('fill', page, { locator: lastLocator, outcome: lastOutcome }, lastActionability, err, {
        field_count: fields.length,
        note: joinNoteTokens(['submit_failed', errorMessage(err)])
      });
    }
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  const notes = [];
  if (fallbackUsed) {
    notes.push('dom_value_fallback');
  }
  if (submitFallbackUsed) {
    notes.push('dom_submit_fallback');
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    field_count: fields.length,
    status: 'filled',
    submitted: submitted,
    note: notes.join(','),
    actionability: lastActionability,
    resolver_outcome: lastOutcome
  };
}

async function selectOptionPage(params) {
  const values = selectValues(params);
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      values: values,
      status: 'selected',
      note: 'mock select'
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const target = await resolveActionLocatorForPage(page, params, { required: true });
  const actionability = await browserActionabilityReportForLocator('select', target.locator, target.outcome, params);
  let selectedValues;
  try {
    selectedValues = await target.locator.selectOption(values);
  } catch (err) {
    return browserActionFailureResult('select', page, target, actionability, err, {
      values
    });
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    values: selectedValues,
    status: 'selected',
    actionability,
    resolver_outcome: target.outcome
  };
}

async function hoverPage(params) {
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      status: 'hovered',
      note: 'mock hover'
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const target = await resolveActionLocatorForPage(page, params, { required: true });
  const actionability = await browserActionabilityReportForLocator('hover', target.locator, target.outcome, params);
  try {
    await target.locator.hover();
  } catch (err) {
    return browserActionFailureResult('hover', page, target, actionability, err);
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    status: 'hovered',
    actionability,
    resolver_outcome: target.outcome
  };
}

async function dragPage(params) {
  if (mockMode) {
    return {
      backend: backendName,
      browser_app: browserApp,
      final_url: stringValue(params.url) || 'about:blank',
      title: 'Mock Page',
      status: 'dragged',
      note: 'mock drag'
    };
  }
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const startTarget = await resolveActionLocatorForPage(page, dragStartLocatorParams(params), { required: true });
  const endTarget = await resolveActionLocatorForPage(page, dragEndLocatorParams(params), { required: true });
  const actionability = await browserActionabilityReportForLocator('drag', startTarget.locator, startTarget.outcome, dragStartLocatorParams(params));
  try {
    await startTarget.locator.dragTo(endTarget.locator);
  } catch (err) {
    return browserActionFailureResult('drag', page, startTarget, actionability, err);
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    status: 'dragged',
    actionability,
    resolver_outcome: dragResolverOutcome(startTarget.outcome, endTarget.outcome)
  };
}

async function clickPage(params) {
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const target = await resolveActionLocatorForPage(page, params, { required: true });
  const timeout = waitTimeout(params);
  const popupIntent = await popupIntentForLocator(target.locator);
  const popupTimeout = popupClickWaitTimeout(timeout);
  const actionability = await browserActionabilityReportForLocator('click', target.locator, target.outcome, params);
  const clickOptions = {};
  const clickTimeout = clickActionTimeout(timeout, popupIntent.likely);
  if (clickTimeout > 0) {
    clickOptions.timeout = clickTimeout;
  }
  if (Boolean(params.force)) {
    clickOptions.force = true;
  }
  if (popupIntent.likely) {
    clickOptions.noWaitAfter = true;
  }
  let popupResult;
  try {
    popupResult = await clickLocatorWithPopupHandling(page, target.locator, clickOptions, popupIntent.likely, popupTimeout);
    applyActionabilityCheckOutcome(actionability, 'navigation_wait', popupResult.navigationWait);
  } catch (err) {
    applyActionabilityCheckOutcome(actionability, 'navigation_wait', err?.navigationWaitOutcome || clickNavigationWaitFailureOutcome(err, clickOptions));
    return browserActionFailureResult('click', page, target, actionability, err);
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  const notes = [];
  if (popupIntent.likely) {
    notes.push(popupResult.popup ? 'popup_opened' : 'popup_not_detected');
  }
  if (popupResult.retried) {
    notes.push('popup_retry');
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    status: 'ok',
    note: notes.join(','),
    actionability,
    resolver_outcome: target.outcome
  };
}

async function typeIntoPage(params) {
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const text = stringValue(params.text);
  const timeout = waitTimeout(params);
  const target = await resolveTypeLocatorForPage(page, params, timeout);
  const locator = target.locator;
  const actionability = await browserActionabilityReportForLocator('type', locator, target.outcome, params);
  if (actionabilityFailedOnAny(actionability, ['enabled', 'editable'])) {
    return browserActionFailureResult('type', page, target, actionability, new Error(fieldActionabilityFailureMessage(actionability)), {
      value: text
    });
  }
  let typeResult;
  try {
    typeResult = await typeLocatorTextWithFallback(locator, text, timeout);
  } catch (err) {
    return browserActionFailureResult('type', page, target, actionability, err, {
      value: text
    });
  }
  let submitted = false;
  let submitFallbackUsed = false;
  if (Boolean(params.submit)) {
    try {
      const submitResult = await submitLocatorWithFallback(locator, timeout);
      submitted = true;
      submitFallbackUsed = Boolean(submitResult.fallbackUsed);
    } catch (err) {
      return browserActionFailureResult('type', page, target, actionability, err, {
        value: text,
        note: joinNoteTokens(['submit_failed', errorMessage(err)])
      });
    }
  }
  if (intValue(params.post_wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.post_wait_ms));
  }
  clearStaleTargetState(page);
  const notes = [];
  if (typeResult.fallbackUsed) {
    notes.push('dom_value_fallback');
  } else if (typeResult.mode === 'keyboard_type') {
    notes.push('keyboard_type');
  }
  if (submitFallbackUsed) {
    notes.push('dom_submit_fallback');
  }
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: page.url(),
    title: await page.title(),
    value: text,
    status: 'ok',
    submitted: submitted,
    note: notes.join(','),
    actionability,
    resolver_outcome: target.outcome
  };
}

async function resolveTypeLocatorForPage(page, params, timeout) {
  try {
    return await resolveActionLocatorForPage(page, params, { required: true });
  } catch (err) {
    if (!timeoutErrorLike(err)) {
      throw err;
    }
    const selector = resolveActionSelector(params);
    if (!selector) {
      throw err;
    }
    const preferred = await preferredSelectorLocator(page, selector, { allowHidden: false });
    const locator = preferred.locator;
    const count = await locator.count().catch(() => 0);
    if (!count) {
      throw err;
    }
    return {
      locator,
      selector,
      outcome: normalizeResolverOutcome({
        status: 'matched',
        resolution_mode: 'selector_fallback_attached',
        primary_kind: 'selector',
        attempt_count: 1,
        matched_kind: 'selector',
        matched_index: preferred.index,
        note: preferred.index > 0
          ? `resolved via selector fallback visible_index=${preferred.index} after visibility wait timeout=${timeout}`
          : `resolved via selector fallback after visibility wait timeout=${timeout}`
      })
    };
  }
}

async function preferredSelectorLocator(page, selector, options = {}) {
  const allowHidden = Boolean(options.allowHidden);
  const matches = page.locator(selector);
  const count = await matches.count().catch(() => 0);
  if (count <= 0 || allowHidden) {
    return { locator: matches.first(), index: 0 };
  }
  const visibleIndex = await page.evaluate((rawSelector) => {
    const elements = Array.from(document.querySelectorAll(rawSelector));
    for (let index = 0; index < elements.length; index += 1) {
      const element = elements[index];
      if (!(element instanceof Element)) {
        continue;
      }
      const style = window.getComputedStyle(element);
      if (!style || style.display === 'none' || style.visibility === 'hidden') {
        continue;
      }
      const rect = element.getBoundingClientRect();
      if (rect.width <= 0 || rect.height <= 0) {
        continue;
      }
      return index;
    }
    return 0;
  }, selector).catch(() => 0);
  const normalizedIndex = Math.max(0, Math.min(count - 1, intValue(visibleIndex)));
  return { locator: matches.nth(normalizedIndex), index: normalizedIndex };
}

async function evalPage(params) {
  const page = await preparePage(params, { allowBlank: false, createPage: true });
  const script = stringValue(params.script);
  if (!script) {
    throw httpError(400, 'script is required');
  }
  const value = await page.evaluate((source) => {
    return globalThis.eval(source);
  }, script);
  clearStaleTargetState(page);
  return {
    Backend: backendName,
    BrowserApp: browserApp,
    FinalURL: page.url(),
    Title: await page.title(),
    Result: truncate(stringifyValue(value), intValue(params.max_chars)),
    Status: 'ok'
  };
}

async function inspectActionResolver(params) {
  if (!enableTestHooks) {
    throw httpError(404, 'test hooks are disabled');
  }
  const page = await preparePage(params, { allowBlank: false, createPage: false });
  const nativeRef = stringValue(params.element_ref);
  const resolver = buildActionResolver(page, params);
  const nativePrimary = nativeRefPrimaryLocatorCandidate(nativeRefEntryForResolver(page, nativeRef));
  return {
    backend: backendName,
    browser_app: browserApp,
    resolver: exportActionResolver(resolver),
    native_ref_primary: nativePrimary ? normalizeLocatorCandidate(nativePrimary) : null
  };
}

async function resolveArtifact(params) {
  const backendPath = stringValue(params.backend_path) || stringValue(params.path);
  if (!backendPath) {
    throw httpError(400, 'backend_path is required');
  }
  const resolved = resolveArtifactPath(backendPath);
  try {
    await fs.access(resolved);
  } catch {
    throw httpError(404, `artifact not found: ${backendPath}`);
  }
  return {
    path: resolved
  };
}

async function preparePage(params, options) {
  if (mockMode) {
    return {
      url() {
        return stringValue(params.url) || 'about:blank';
      },
      async title() {
        return 'Mock Page';
      }
    };
  }
  await ensureBrowser(requestedProfile(params));
  const page = await selectPage(params, options);
  const url = stringValue(params.url);
  if (url) {
    await gotoPageWithLoadFallback(page, url, waitTimeout(params));
  } else if (!options.allowBlank && page.url() === 'about:blank') {
    throw httpError(400, 'url is required for this action until a page has been opened');
  }
  if (intValue(params.wait_ms) > 0) {
    await page.waitForTimeout(intValue(params.wait_ms));
  }
  runtimeState.activePage = page;
  return page;
}

async function selectPage(params, options) {
  const pages = currentPages();
  const tabIndex = intValue(params.tab_index);
  if (tabIndex > 0) {
    return pageByTabIndex(pages, tabIndex);
  }
  if (runtimeState.activePage && !runtimeState.activePage.isClosed()) {
    return runtimeState.activePage;
  }
  if (pages.length > 0) {
    return pages[0];
  }
  if (!options.createPage) {
    throw httpError(400, 'no active page');
  }
  return runtimeState.context.newPage();
}

async function resolveActionLocatorForPage(page, params, options = {}) {
  const timeout = waitTimeout(params);
  const resolver = buildActionResolver(page, params);
  if (resolver) {
    return resolveActionLocatorWithResolver(page, resolver, params, timeout, options);
  }
  const selector = resolveActionSelector(params);
  if (!selector) {
    if (options.required) {
      throw httpError(400, 'selector or element_ref is required');
    }
    return { locator: null, selector: '', outcome: null };
  }
  const preferred = await preferredSelectorLocator(page, selector, { allowHidden: options.allowHidden });
  const locator = preferred.locator;
  try {
    await locator.waitFor({ state: options.allowHidden ? 'attached' : 'visible', timeout });
  } catch (err) {
    if (options.allowHidden || !timeoutErrorLike(err)) {
      throw err;
    }
    const count = await locator.count().catch(() => 0);
    if (!count) {
      throw err;
    }
    return {
      locator,
      selector,
      outcome: actionTargetRequested(params)
        ? normalizeResolverOutcome({
          status: 'matched',
          resolution_mode: 'selector_fallback_attached',
          primary_kind: 'selector',
          attempt_count: 1,
          matched_kind: 'selector',
          matched_index: preferred.index,
          note: preferred.index > 0
            ? `resolved via selector fallback visible_index=${preferred.index} after visibility wait timeout=${timeout}`
            : `resolved via selector fallback after visibility wait timeout=${timeout}`
        })
        : null
    };
  }
  return {
    locator,
    selector,
    outcome: actionTargetRequested(params)
      ? normalizeResolverOutcome({
        status: 'matched',
        resolution_mode: 'selector_first',
        primary_kind: 'selector',
        attempt_count: 1,
        matched_kind: 'selector',
        matched_index: preferred.index,
        note: preferred.index > 0 ? `resolved via selector visible_index=${preferred.index}` : 'resolved via selector'
      })
      : null
  };
}

async function ensureBrowser(profileName, options = {}) {
  const desiredProfile = ensureProfile(profileName || runtimeState.activeProfile);
  enforceAttachProfile(desiredProfile);
  const allowEscalatedRestart = options && options.allowEscalatedRestart === true;
  if (!runtimeState.browser) {
    if (!allowEscalatedRestart) {
      enforcePermanentRestartFailure(desiredProfile);
      enforceDisconnectCooldown(desiredProfile);
      enforceRestartRetryBackoff(desiredProfile);
    }
    const restartAttempted = beginDisconnectRestartAttempt(desiredProfile);
    try {
      if (restartAttempted) {
        await maybeInjectRestartFailure(desiredProfile);
      }
      const playwright = await loadPlaywright();
      runtimeState.browser = attachCDPMode
        ? await playwright.chromium.connectOverCDP(attachCDPEndpoint)
        : await playwright.chromium.launch({ headless: true });
      attachBrowserLifecycle(runtimeState.browser);
      if (restartAttempted) {
        completeDisconnectRestartAttempt(desiredProfile, 'succeeded', '');
      }
    } catch (err) {
      if (restartAttempted) {
        completeDisconnectRestartAttempt(desiredProfile, 'failed', errorMessage(err));
      }
      throw err;
    }
  }
  if (attachCDPMode) {
    await bindAttachedBrowserContext(desiredProfile);
    return;
  }
  if (runtimeState.context && normalizedProfile(runtimeState.activeProfile) === desiredProfile) {
    return;
  }
  if (runtimeState.context) {
    await persistActiveProfileState();
    await closeActiveProfileContext({ closeBrowser: false });
  }
  const contextOptions = await browserContextOptionsForProfile(desiredProfile);
  runtimeState.context = await runtimeState.browser.newContext(contextOptions);
  runtimeState.activeProfile = desiredProfile;
  runtimeState.activePage = runtimeState.context.pages().find((page) => !page.isClosed()) || null;
  clearBrowserDisconnectStateForProfile(desiredProfile);
  runtimeState.context.on('page', (page) => {
    registerManagedPage(page);
    runtimeState.activePage = page;
  });
  currentPages().forEach((page) => registerManagedPage(page));
}

async function closeBrowser() {
  await persistActiveProfileState();
  await closeActiveProfileContext({ closeBrowser: false });
  if (attachCDPMode) {
    runtimeState.browser = null;
    return;
  }
  if (runtimeState.browser) {
    runtimeState.browserCloseIntent = true;
    try {
      await runtimeState.browser.close().catch(() => {});
    } finally {
      runtimeState.browserCloseIntent = false;
    }
  }
  runtimeState.browser = null;
}

async function loadPlaywright() {
  if (mockMode) {
    return null;
  }
  if (runtimeState.playwright) {
    return runtimeState.playwright;
  }
  for (const packageName of ['playwright', 'playwright-core']) {
    try {
      const loaded = await import(packageName);
      runtimeState.playwright = loaded;
      runtimeState.playwrightPackage = packageName;
      return loaded;
    } catch {
      continue;
    }
  }
  throw httpError(
    503,
    'playwright is not installed for agentx-browserd; run npm install in core/agentx/browserdaemon/node'
  );
}

function currentPages() {
  if (!runtimeState.context) {
    return [];
  }
  return runtimeState.context.pages().filter((page) => !page.isClosed());
}

function attachBrowserLifecycle(browser) {
  if (!browser || typeof browser.on !== 'function') {
    return;
  }
  browser.on('disconnected', () => {
    handleBrowserDisconnected();
  });
}

function handleBrowserDisconnected() {
  if (runtimeState.browserCloseIntent) {
    return;
  }
  const profile = normalizedProfile(runtimeState.activeProfile);
  const retained = cloneSessionState(currentSessionStateForProfile(profile)) || {
    pendingDialog: null,
    pendingUpload: null,
    pendingDownloads: [],
    downloadWaiters: [],
    nextDownloadSeq: 0,
    staleTarget: null,
    browserDisconnect: null,
    disconnectHistory: null,
    lastDownload: null,
    lastDialog: null,
    events: [],
    lastLifecycleEvent: '',
    popupCount: 0,
    updatedAt: 0
  };
  const disconnectHistory = disconnectHistoryForState(retained);
  const disconnectCount = intValue(disconnectHistory.count) + 1;
  const disconnectedAt = Date.now();
  const recentOccurredAt = normalizedDisconnectRecentOccurredAt(disconnectHistory.recentOccurredAt, disconnectedAt);
  recentOccurredAt.push(disconnectedAt);
  const burstCount = recentOccurredAt.length;
  const recommendedBackoffMs = disconnectBackoffMsForCount(disconnectCount);
  const reconnectHint = disconnectReconnectHintForBurstCount(burstCount);
  retained.disconnectHistory = {
    ...disconnectHistory,
    count: disconnectCount,
    lastOccurredAt: disconnectedAt,
    lastRecoveredAt: intValue(disconnectHistory.lastRecoveredAt),
    recommendedBackoffMs,
    recentOccurredAt
  };
  retained.browserDisconnect = {
    message: 'browser runtime disconnected',
    recoveryAction: disconnectRecoveryActionForBurstCount(burstCount),
    occurredAt: disconnectedAt,
    count: disconnectCount,
    burstCount,
    burstWindowMs: disconnectBurstWindowMs,
    recommendedBackoffMs,
    reconnectHint
  };
  retained.updatedAt = disconnectedAt;
  runtimeState.retainedSessionStateByProfile.set(profile, retained);
  runtimeState.browser = null;
  runtimeState.context = null;
  runtimeState.activePage = null;
  runtimeState.pageNativeRefs.clear();
  runtimeState.pageSessionState.clear();
  runtimeState.nextNativeRefID = 1;
  runtimeState.managedPages = new WeakSet();
}

function pageByTabIndex(pages, index) {
  if (index <= 0 || index > pages.length) {
    throw httpError(400, `tab_index out of range: ${index}`);
  }
  return pages[index - 1];
}

function currentTabIndexForPage(page) {
  if (!page || page.isClosed()) {
    return 0;
  }
  const pages = currentPages();
  const index = pages.indexOf(page);
  if (index < 0) {
    return 0;
  }
  return index + 1;
}

function ensureProfile(profileName) {
  const profile = normalizedProfile(profileName);
  const existing = runtimeState.profiles.get(profile) || {};
  runtimeState.profiles.set(profile, {
    ...existing,
    profile,
    root: path.join(profilesRoot, profile),
    storage_state_path: path.join(profilesRoot, profile, 'storage-state.json')
  });
  return profile;
}

function normalizedProfile(profileName) {
  return stringValue(profileName) || defaultProfile;
}

function enforceAttachProfile(profileName) {
  if (attachCDPMode && normalizedProfile(profileName) !== defaultProfile) {
    throw httpError(400, 'host CDP attach only supports the default profile');
  }
}

function requestedProfile(params) {
  return normalizedProfile(stringValue(params?.profile) || runtimeState.activeProfile);
}

function profileRecord(profileName) {
  const profile = ensureProfile(profileName);
  return runtimeState.profiles.get(profile);
}

async function ensureProfileDir(profileName) {
  const record = profileRecord(profileName);
  await fs.mkdir(record.root, { recursive: true });
  return record;
}

async function browserContextOptionsForProfile(profileName) {
  const record = await ensureProfileDir(profileName);
  const out = {
    acceptDownloads: true
  };
  try {
    const raw = await fs.readFile(record.storage_state_path, 'utf8');
    if (stringValue(raw)) {
      out.storageState = record.storage_state_path;
    }
  } catch {
    return out;
  }
  return out;
}

async function persistActiveProfileState() {
  if (attachCDPMode) {
    return;
  }
  if (!runtimeState.context) {
    return;
  }
  const record = await ensureProfileDir(runtimeState.activeProfile);
  await runtimeState.context.storageState({ path: record.storage_state_path }).catch(() => {});
}

async function closeActiveProfileContext(options = {}) {
  if (attachCDPMode) {
    runtimeState.context = null;
    runtimeState.activePage = null;
    runtimeState.pageNativeRefs.clear();
    runtimeState.pageSessionState.clear();
    runtimeState.retainedSessionStateByProfile.clear();
    runtimeState.nextNativeRefID = 1;
    runtimeState.managedPages = new WeakSet();
    if (options.closeBrowser) {
      runtimeState.browser = null;
    }
    return;
  }
  if (runtimeState.context) {
    await runtimeState.context.close().catch(() => {});
  }
  runtimeState.context = null;
  runtimeState.activePage = null;
  runtimeState.pageNativeRefs.clear();
  runtimeState.pageSessionState.clear();
  runtimeState.retainedSessionStateByProfile.clear();
  runtimeState.nextNativeRefID = 1;
  runtimeState.managedPages = new WeakSet();
  if (options.closeBrowser && runtimeState.browser) {
    runtimeState.browserCloseIntent = true;
    try {
      await runtimeState.browser.close().catch(() => {});
    } finally {
      runtimeState.browserCloseIntent = false;
    }
    runtimeState.browser = null;
  }
}

async function bindAttachedBrowserContext(profileName) {
  if (!runtimeState.browser) {
    throw httpError(503, 'browser attach is not connected');
  }
  const profile = ensureProfile(profileName);
  enforceAttachProfile(profile);
  const contexts = typeof runtimeState.browser.contexts === 'function'
    ? runtimeState.browser.contexts().filter((context) => context && typeof context.pages === 'function')
    : [];
  let context = contexts[0] || null;
  if (!context && typeof runtimeState.browser.newContext === 'function') {
    context = await runtimeState.browser.newContext({ acceptDownloads: true });
  }
  if (!context) {
    throw httpError(503, 'browser attach did not expose a usable context');
  }
  if (runtimeState.context === context && normalizedProfile(runtimeState.activeProfile) === profile) {
    return;
  }
  runtimeState.context = context;
  runtimeState.activeProfile = profile;
  runtimeState.activePage = runtimeState.context.pages().find((page) => !page.isClosed()) || null;
  clearBrowserDisconnectStateForProfile(profile);
  runtimeState.context.on('page', (page) => {
    registerManagedPage(page);
    runtimeState.activePage = page;
  });
  currentPages().forEach((page) => registerManagedPage(page));
}

async function copyProfileStorageState(copyFromProfile, targetProfile) {
  const source = profileRecord(copyFromProfile);
  const target = await ensureProfileDir(targetProfile);
  if (!source || stringValue(source.storage_state_path) === '' || source.profile === target.profile) {
    return;
  }
  try {
    await fs.access(source.storage_state_path);
  } catch {
    return;
  }
  await fs.copyFile(source.storage_state_path, target.storage_state_path).catch(() => {});
}

function isProfileRunning(profileName) {
  return Boolean(runtimeState.context) && normalizedProfile(runtimeState.activeProfile) === normalizedProfile(profileName);
}

function resolveArtifactPath(artifactPathValue) {
  if (path.isAbsolute(artifactPathValue)) {
    return artifactPathValue;
  }
  return path.join(artifactsRoot, artifactPathValue);
}

function artifactPath(prefix, extension) {
  return path.join(artifactsRoot, `${prefix}-${timestampTag()}-${randomUUID()}.${extension}`);
}

async function writeMockScreenshot() {
  const targetPath = artifactPath('mock-screenshot', 'png');
  const blob = Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO2kK5QAAAAASUVORK5CYII=',
    'base64'
  );
  await fs.writeFile(targetPath, blob);
  return targetPath;
}

async function writeMockDownload(requestedPath) {
  const targetPath = requestedPath || artifactPath('mock-download', 'zip');
  await fs.mkdir(path.dirname(targetPath), { recursive: true });
  const blob = Buffer.from('PK\x03\x04agentx-browserd-mock-download', 'utf8');
  await fs.writeFile(targetPath, blob);
  return targetPath;
}

function targetIDForPage(page) {
  let value = pageTargetIDs.get(page);
  if (!value) {
    value = `page-${randomUUID()}`;
    pageTargetIDs.set(page, value);
  }
  return value;
}

function registerManagedPage(page) {
  if (!page || runtimeState.managedPages.has(page)) {
    return;
  }
  runtimeState.managedPages.add(page);
  pageSessionStateForPage(page);
  page.on('framenavigated', (frame) => {
    try {
      if (frame === page.mainFrame()) {
        recordPageLifecycle(page, 'main_frame_navigated');
      }
    } catch {
      recordPageLifecycle(page, 'main_frame_navigated');
    }
  });
  page.on('popup', (popup) => {
    recordPagePopup(page, popup);
    registerManagedPage(popup);
  });
  page.on('download', (download) => {
    handlePageDownload(page, download);
  });
  page.on('dialog', async (dialog) => {
    await handlePageDialog(page, dialog);
  });
  page.on('filechooser', async (chooser) => {
    await handlePageFileChooser(page, chooser);
  });
  page.on('pageerror', (err) => {
    recordPageError(page, err);
  });
  page.on('close', () => {
    clearNativeRefsForPage(page);
    recordPageLifecycle(page, 'closed');
    clearPageSessionState(page);
    if (runtimeState.activePage === page) {
      runtimeState.activePage = currentPages()[0] || null;
    }
  });
}

function clearNativeRefsForPage(page) {
  const targetID = pageTargetIDs.get(page);
  if (!targetID) {
    return;
  }
  runtimeState.pageNativeRefs.delete(targetID);
}

function pageSessionStateForPage(page) {
  const targetID = targetIDForPage(page);
  let state = runtimeState.pageSessionState.get(targetID);
  if (!state) {
    const retained = retainedSessionStateForProfile(runtimeState.activeProfile);
    state = {
      pendingDialog: null,
      pendingUpload: null,
      pendingDownloads: [],
      downloadWaiters: [],
      nextDownloadSeq: 0,
      staleTarget: null,
      browserDisconnect: null,
      disconnectHistory: retained && typeof retained === 'object' && retained.disconnectHistory
        ? {
          ...objectValue(retained.disconnectHistory),
          recentOccurredAt: normalizedDisconnectRecentOccurredAt(objectValue(retained.disconnectHistory).recentOccurredAt)
        }
        : null,
      lastDownload: null,
      lastDialog: null,
      events: [],
      lastLifecycleEvent: '',
      popupCount: 0,
      updatedAt: 0
    };
    runtimeState.pageSessionState.set(targetID, state);
  }
  return state;
}

function clearPageSessionState(page) {
  const targetID = pageTargetIDs.get(page);
  if (!targetID) {
    return;
  }
  const state = runtimeState.pageSessionState.get(targetID);
  if (state) {
    retainSessionStateForProfile(runtimeState.activeProfile, state);
  }
  if (state && Array.isArray(state.downloadWaiters)) {
    for (const waiter of state.downloadWaiters) {
      try {
        waiter.reject(httpError(409, 'page closed before download completed'));
      } catch {
        continue;
      }
    }
  }
  runtimeState.pageSessionState.delete(targetID);
}

function recordPageLifecycle(page, eventName) {
  const state = pageSessionStateForPage(page);
  state.lastLifecycleEvent = stringValue(eventName);
  if (state.lastLifecycleEvent === 'main_frame_navigated' || state.lastLifecycleEvent === 'closed') {
    state.staleTarget = null;
    state.pendingUpload = null;
  }
  state.updatedAt = Date.now();
  if (state.lastLifecycleEvent === 'main_frame_navigated' || state.lastLifecycleEvent === 'closed' || state.lastLifecycleEvent === 'download_started') {
    pushPageEvent(page, 'page_lifecycle', state.lastLifecycleEvent, lifecycleEventMessage(state.lastLifecycleEvent), safeCall(() => page.url()));
  }
}

function recordPagePopup(page, popup) {
  const state = pageSessionStateForPage(page);
  state.lastLifecycleEvent = 'popup_opened';
  state.popupCount += 1;
  state.updatedAt = Date.now();
  pushPageEvent(page, 'popup', 'popup_opened', 'popup opened', safeCall(() => popup.url()));
}

function recordPageError(page, err) {
  const message = errorMessage(err);
  if (!message) {
    return;
  }
  pushPageEvent(page, 'pageerror', 'page_error', message, safeCall(() => page.url()));
}

function lifecycleEventMessage(eventName) {
  switch (stringValue(eventName)) {
    case 'main_frame_navigated':
      return 'main frame navigated';
    case 'download_started':
      return 'download started';
    case 'closed':
      return 'page closed';
    default:
      return stringValue(eventName);
  }
}

function pushPageEvent(page, source, event, message, url, details = null) {
  const state = pageSessionStateForPage(page);
  const currentDetails = objectValue(details);
  const metadata = pageEventMetadata(source, event);
  const entry = {
    event: stringValue(event),
    category: stringValue(metadata.category),
    severity: stringValue(metadata.severity),
    resolverStatus: stringValue(currentDetails.resolver_status || currentDetails.resolverStatus),
    candidateKind: stringValue(currentDetails.candidate_kind || currentDetails.candidateKind),
    candidateStrength: stringValue(currentDetails.candidate_strength || currentDetails.candidateStrength),
    ambiguityClass: stringValue(currentDetails.ambiguity_class || currentDetails.ambiguityClass),
    retryDisposition: stringValue(currentDetails.retry_disposition || currentDetails.retryDisposition),
    manualRetryHint: stringValue(currentDetails.manual_retry_hint || currentDetails.manualRetryHint),
    nextStepAlias: stringValue(currentDetails.next_step_alias || currentDetails.nextStepAlias),
    blockedBy: stringValue(currentDetails.blocked_by || currentDetails.blockedBy),
    locatorCount: intValue(currentDetails.locator_count || currentDetails.locatorCount),
    candidateCount: intValue(currentDetails.candidate_count || currentDetails.candidateCount),
    preferredOrdinal: intValue(currentDetails.preferred_ordinal || currentDetails.preferredOrdinal),
    specificityFields: stringSliceValue(currentDetails.specificity_fields || currentDetails.specificityFields),
    recoveryAction: stringValue(metadata.recoveryAction),
    targetID: targetIDForPage(page),
    tabIndex: currentTabIndexForPage(page),
    message: stringValue(message),
    source: stringValue(source),
    url: stringValue(url),
    occurredAt: Date.now()
  };
  if (!entry.message || !entry.source || !entry.event) {
    return;
  }
  state.events.push(entry);
  if (state.events.length > sessionEventBufferLimit) {
    state.events.splice(0, state.events.length - sessionEventBufferLimit);
  }
  state.updatedAt = Date.now();
}

function browserErrorEntriesForState(state) {
  return arrayValue(state?.events).map((entry) => ({
    event: stringValue(entry.event),
    category: stringValue(entry.category),
    severity: stringValue(entry.severity),
    resolver_status: stringValue(entry.resolverStatus),
    candidate_kind: stringValue(entry.candidateKind),
    candidate_strength: stringValue(entry.candidateStrength),
    ambiguity_class: stringValue(entry.ambiguityClass),
    retry_disposition: stringValue(entry.retryDisposition),
    manual_retry_hint: stringValue(entry.manualRetryHint),
    next_step_alias: stringValue(entry.nextStepAlias),
    blocked_by: stringValue(entry.blockedBy),
    locator_count: intValue(entry.locatorCount),
    candidate_count: intValue(entry.candidateCount),
    preferred_ordinal: intValue(entry.preferredOrdinal),
    specificity_fields: stringSliceValue(entry.specificityFields),
    recovery_action: stringValue(entry.recoveryAction),
    target_id: stringValue(entry.targetID),
    tab_index: intValue(entry.tabIndex),
    message: stringValue(entry.message),
    source: stringValue(entry.source),
    url: stringValue(entry.url),
    occurred_at: intValue(entry.occurredAt)
  }));
}

function pageEventMetadata(source, event) {
  const currentEvent = stringValue(event);
  const currentSource = stringValue(source);
  switch (currentEvent) {
    case 'popup_opened':
      return { category: 'popup', severity: 'info', recoveryAction: '' };
    case 'main_frame_navigated':
      return { category: 'navigation', severity: 'info', recoveryAction: '' };
    case 'download_started':
      return { category: 'artifact', severity: 'info', recoveryAction: '' };
    case 'closed':
      return { category: 'page_lifecycle', severity: 'warn', recoveryAction: 'browser action=ensure' };
    case 'stale_target':
      return { category: 'resolver', severity: 'warn', recoveryAction: 'browser action=snapshot' };
    case 'ambiguous_target':
      return { category: 'resolver', severity: 'info', recoveryAction: 'browser action=snapshot' };
    case 'page_error':
      return { category: 'script', severity: 'error', recoveryAction: 'browser action=refresh' };
    default:
      if (currentSource === 'pageerror') {
        return { category: 'script', severity: 'error', recoveryAction: 'browser action=refresh' };
      }
      return { category: currentSource || 'event', severity: 'info', recoveryAction: '' };
  }
}

function currentPageForProfile(profileName) {
  if (!runtimeState.context || normalizedProfile(runtimeState.activeProfile) !== normalizedProfile(profileName)) {
    return null;
  }
  if (runtimeState.activePage && !runtimeState.activePage.isClosed()) {
    return runtimeState.activePage;
  }
  return currentPages()[0] || null;
}

function currentSessionStateForProfile(profileName) {
  const page = currentPageForProfile(profileName);
  if (page) {
    return pageSessionStateForPage(page);
  }
  return retainedSessionStateForProfile(profileName);
}

function retainedSessionStateForProfile(profileName) {
  const profile = normalizedProfile(profileName);
  return runtimeState.retainedSessionStateByProfile.get(profile) || null;
}

function retainSessionStateForProfile(profileName, state) {
  const profile = normalizedProfile(profileName);
  const snapshot = cloneSessionState(state);
  if (!snapshot) {
    runtimeState.retainedSessionStateByProfile.delete(profile);
    return;
  }
  runtimeState.retainedSessionStateByProfile.set(profile, snapshot);
}

function cloneSessionState(state) {
  const current = objectValue(state);
  if (Object.keys(current).length === 0) {
    return null;
  }
  return {
    pendingDialog: current.pendingDialog ? { ...objectValue(current.pendingDialog) } : null,
    pendingUpload: current.pendingUpload ? {
      ...objectValue(current.pendingUpload),
      paths: arrayValue(current.pendingUpload.paths).map((item) => stringValue(item)).filter(Boolean)
    } : null,
    pendingDownloads: arrayValue(current.pendingDownloads).map((entry) => ({ ...objectValue(entry) })),
    downloadWaiters: [],
    nextDownloadSeq: intValue(current.nextDownloadSeq),
    staleTarget: current.staleTarget ? { ...objectValue(current.staleTarget) } : null,
    browserDisconnect: current.browserDisconnect ? { ...objectValue(current.browserDisconnect) } : null,
    disconnectHistory: current.disconnectHistory ? {
      ...objectValue(current.disconnectHistory),
      recentOccurredAt: normalizedDisconnectRecentOccurredAt(objectValue(current.disconnectHistory).recentOccurredAt)
    } : null,
    lastDownload: current.lastDownload ? { ...objectValue(current.lastDownload) } : null,
    lastDialog: current.lastDialog ? { ...objectValue(current.lastDialog) } : null,
    events: arrayValue(current.events).map((entry) => ({ ...objectValue(entry) })),
    lastLifecycleEvent: stringValue(current.lastLifecycleEvent),
    popupCount: intValue(current.popupCount),
    updatedAt: intValue(current.updatedAt)
  };
}

function browserStatusNoteTokens(profileName) {
  const tokens = [];
  if (mockMode) {
    tokens.push('runtime_mode=mock');
  } else if (runtimeState.playwrightPackage) {
    tokens.push(noteToken('playwright_package', runtimeState.playwrightPackage));
  } else {
    tokens.push('daemon_status=ready');
  }
  const cache = playwrightCacheSummary();
  if (cache) {
    if (cache.host_os) {
      tokens.push(noteToken('playwright_host_os', cache.host_os));
    }
    if (cache.host_arch) {
      tokens.push(noteToken('playwright_host_arch', cache.host_arch));
    }
    if (cache.node_version) {
      tokens.push(noteToken('playwright_node_version', cache.node_version));
    }
    if (cache.playwright_package) {
      tokens.push(noteToken('playwright_package', cache.playwright_package));
    }
    if (cache.playwright_package_version) {
      tokens.push(noteToken('playwright_package_version', cache.playwright_package_version));
    }
    if (cache.runtime_summary_generation) {
      tokens.push(noteToken('playwright_runtime_summary_generation', cache.runtime_summary_generation));
    }
    tokens.push(noteToken('playwright_runtime_baseline_ready', cache.runtime_baseline_ready ? 'true' : 'false'));
    if (cache.runtime_baseline_block_reason) {
      tokens.push(noteToken('playwright_runtime_baseline_block_reason', cache.runtime_baseline_block_reason));
    }
    if (cache.source) {
      tokens.push(noteToken('playwright_cache_source', cache.source));
    }
    tokens.push(noteToken('playwright_cache_pinned', cache.pinned ? 'true' : 'false'));
    if (cache.policy_version) {
      tokens.push(noteToken('playwright_cache_policy', cache.policy_version));
    }
    if (cache.retention_mode) {
      tokens.push(noteToken('playwright_cache_retention', cache.retention_mode));
    }
    if (cache.bootstrap_state) {
      tokens.push(noteToken('playwright_bootstrap_state', cache.bootstrap_state));
    }
    if (cache.bootstrap_error_code) {
      tokens.push(noteToken('playwright_bootstrap_error', cache.bootstrap_error_code));
    }
    if (cache.browser_revision) {
      tokens.push(noteToken('playwright_browser_revision', cache.browser_revision));
    }
    if (cache.delivery_generation) {
      tokens.push(noteToken('playwright_delivery_generation', cache.delivery_generation));
    }
    if (cache.target_delivery_generation) {
      tokens.push(noteToken('playwright_target_delivery_generation', cache.target_delivery_generation));
    }
    if (cache.last_ready_delivery_generation) {
      tokens.push(noteToken('playwright_last_ready_delivery_generation', cache.last_ready_delivery_generation));
    }
    if (cache.last_evicted_delivery_generation) {
      tokens.push(noteToken('playwright_last_evicted_delivery', cache.last_evicted_delivery_generation));
    }
    if (cache.last_delivery_generation_switch_unix_milli > 0) {
      tokens.push(noteToken('playwright_last_delivery_switch_unix', String(cache.last_delivery_generation_switch_unix_milli)));
    }
    if (cache.retained_delivery_browser_revision) {
      tokens.push(noteToken('playwright_retained_delivery_browser_revision', cache.retained_delivery_browser_revision));
    }
    tokens.push(noteToken('playwright_retained_delivery_cache_ready', cache.retained_delivery_cache_ready ? 'true' : 'false'));
    if (cache.retained_fallback_delivery_generation) {
      tokens.push(noteToken('playwright_retained_fallback_delivery_generation', cache.retained_fallback_delivery_generation));
    }
    tokens.push(noteToken('playwright_retained_fallback_payload_ready', cache.retained_fallback_payload_ready ? 'true' : 'false'));
    if (cache.retained_fallback_payload_block_reason) {
      tokens.push(noteToken('playwright_retained_fallback_payload_block_reason', cache.retained_fallback_payload_block_reason));
    }
    if (cache.retained_fallback_payload_source) {
      tokens.push(noteToken('playwright_retained_fallback_payload_source', cache.retained_fallback_payload_source));
    }
    tokens.push(noteToken('playwright_retained_fallback_launch_ready', cache.retained_fallback_launch_ready ? 'true' : 'false'));
    if (cache.retained_fallback_launch_block_reason) {
      tokens.push(noteToken('playwright_retained_fallback_launch_block_reason', cache.retained_fallback_launch_block_reason));
    }
    tokens.push(noteToken('playwright_selected_launch_ready', cache.selected_launch_ready ? 'true' : 'false'));
    if (cache.selected_launch_source) {
      tokens.push(noteToken('playwright_selected_launch_source', cache.selected_launch_source));
    }
    if (cache.selected_launch_delivery_generation) {
      tokens.push(noteToken('playwright_selected_launch_delivery_generation', cache.selected_launch_delivery_generation));
    }
    if (cache.selected_launch_browser_revision) {
      tokens.push(noteToken('playwright_selected_launch_browser_revision', cache.selected_launch_browser_revision));
    }
    if (cache.selected_launch_payload_source) {
      tokens.push(noteToken('playwright_selected_launch_payload_source', cache.selected_launch_payload_source));
    }
    tokens.push(noteToken('playwright_selected_launch_payload_ready', cache.selected_launch_payload_ready ? 'true' : 'false'));
    if (cache.selected_launch_payload_block_reason) {
      tokens.push(noteToken('playwright_selected_launch_payload_block_reason', cache.selected_launch_payload_block_reason));
    }
    tokens.push(noteToken('playwright_selected_launch_executable_ready', cache.selected_launch_executable_ready ? 'true' : 'false'));
    if (cache.selected_launch_executable_block_reason) {
      tokens.push(noteToken('playwright_selected_launch_executable_block_reason', cache.selected_launch_executable_block_reason));
    }
    if (cache.selected_launch_block_reason) {
      tokens.push(noteToken('playwright_selected_launch_block_reason', cache.selected_launch_block_reason));
    }
    if (cache.delivery_transition_pending) {
      tokens.push(noteToken('playwright_delivery_transition_pending', 'true'));
    }
    if (cache.delivery_transition_stage) {
      tokens.push(noteToken('playwright_delivery_transition_stage', cache.delivery_transition_stage));
    }
    tokens.push(noteToken('playwright_launch_ready', cache.launch_ready ? 'true' : 'false'));
    if (cache.launch_block_reason) {
      tokens.push(noteToken('playwright_launch_block_reason', cache.launch_block_reason));
    }
    tokens.push(noteToken('playwright_bundle_ready', cache.bundle_ready ? 'true' : 'false'));
    tokens.push(noteToken('playwright_delivery_ready', cache.delivery_ready ? 'true' : 'false'));
  }
  appendSessionDiagnosticsTokens(tokens, currentSessionStateForProfile(profileName));
  return tokens;
}

function playwrightCacheSummary() {
  if (!playwrightBrowsersPath && !playwrightCacheSource && !playwrightCachePolicyVersion && !playwrightCacheRetentionMode) {
    return null;
  }
  const runtimeObservedLaunch = runtimeObservedSelectedLaunchSummary();
  const summary = {
    host_os: normalizedHostOS(),
    host_arch: normalizedHostArch(),
    node_version: normalizedNodeVersion(),
    path: playwrightBrowsersPath,
    source: playwrightCacheSource,
    pinned: playwrightCachePinned
  };
  const packageName = resolvedPlaywrightPackageName();
  if (packageName) {
    summary.playwright_package = packageName;
  }
  const packageVersion = resolvedPlaywrightPackageVersion(packageName);
  if (packageVersion) {
    summary.playwright_package_version = packageVersion;
  }
  if (playwrightBundleGeneration) {
    summary.bundle_generation = playwrightBundleGeneration;
  }
  if (playwrightDependencyGeneration) {
    summary.dependency_generation = playwrightDependencyGeneration;
  }
  if (playwrightBrowserRevision) {
    summary.browser_revision = playwrightBrowserRevision;
  }
  if (playwrightDeliveryGeneration) {
    summary.delivery_generation = playwrightDeliveryGeneration;
  }
  if (playwrightTargetDeliveryGeneration) {
    summary.target_delivery_generation = playwrightTargetDeliveryGeneration;
  }
  if (playwrightLastReadyDeliveryGeneration) {
    summary.last_ready_delivery_generation = playwrightLastReadyDeliveryGeneration;
  }
  if (playwrightRetainedDeliveries.length > 0) {
    summary.retained_delivery_generations = playwrightRetainedDeliveries;
  }
  if (playwrightLastEvictedDeliveryGeneration) {
    summary.last_evicted_delivery_generation = playwrightLastEvictedDeliveryGeneration;
  }
  if (playwrightLastDeliverySwitchUnix > 0) {
    summary.last_delivery_generation_switch_unix_milli = playwrightLastDeliverySwitchUnix;
  }
  if (playwrightRetainedDeliveryRevision) {
    summary.retained_delivery_browser_revision = playwrightRetainedDeliveryRevision;
  }
  summary.retained_delivery_cache_ready = playwrightRetainedDeliveryReady;
  if (playwrightRetainedFallbackDeliveryGeneration) {
    summary.retained_fallback_delivery_generation = playwrightRetainedFallbackDeliveryGeneration;
  }
  summary.retained_fallback_payload_ready = playwrightRetainedFallbackPayloadReady;
  if (playwrightRetainedFallbackPayloadBlockReason) {
    summary.retained_fallback_payload_block_reason = playwrightRetainedFallbackPayloadBlockReason;
  }
  if (playwrightRetainedFallbackPayloadSource) {
    summary.retained_fallback_payload_source = playwrightRetainedFallbackPayloadSource;
  }
  if (playwrightRetainedFallbackPayloadDirs.length > 0) {
    summary.retained_fallback_payload_dirs = playwrightRetainedFallbackPayloadDirs;
  }
  summary.retained_fallback_launch_ready = playwrightRetainedFallbackLaunchReady;
  if (playwrightRetainedFallbackLaunchBlockReason) {
    summary.retained_fallback_launch_block_reason = playwrightRetainedFallbackLaunchBlockReason;
  }
  if (playwrightSelectedLaunchDeliveryGeneration) {
    summary.selected_launch_delivery_generation = playwrightSelectedLaunchDeliveryGeneration;
  } else if (runtimeObservedLaunch.delivery_generation) {
    summary.selected_launch_delivery_generation = runtimeObservedLaunch.delivery_generation;
  }
  if (playwrightSelectedLaunchSource) {
    summary.selected_launch_source = playwrightSelectedLaunchSource;
  } else if (runtimeObservedLaunch.source) {
    summary.selected_launch_source = runtimeObservedLaunch.source;
  }
  summary.selected_launch_ready = playwrightSelectedLaunchReady || runtimeObservedLaunch.ready;
  if (playwrightSelectedLaunchBrowserRevision) {
    summary.selected_launch_browser_revision = playwrightSelectedLaunchBrowserRevision;
  } else if (runtimeObservedLaunch.browser_revision) {
    summary.selected_launch_browser_revision = runtimeObservedLaunch.browser_revision;
  }
  if (playwrightSelectedLaunchPayloadSource) {
    summary.selected_launch_payload_source = playwrightSelectedLaunchPayloadSource;
  } else if (runtimeObservedLaunch.payload_source) {
    summary.selected_launch_payload_source = runtimeObservedLaunch.payload_source;
  }
  if (playwrightSelectedLaunchPayloadDirs.length > 0) {
    summary.selected_launch_payload_dirs = playwrightSelectedLaunchPayloadDirs;
  } else if (runtimeObservedLaunch.payload_dirs.length > 0) {
    summary.selected_launch_payload_dirs = runtimeObservedLaunch.payload_dirs;
  }
  summary.selected_launch_payload_ready = playwrightSelectedLaunchPayloadReady || runtimeObservedLaunch.payload_ready;
  if (playwrightSelectedLaunchPayloadBlockReason) {
    summary.selected_launch_payload_block_reason = playwrightSelectedLaunchPayloadBlockReason;
  }
  if (playwrightSelectedLaunchExecutablePath) {
    summary.selected_launch_executable_path = playwrightSelectedLaunchExecutablePath;
  } else if (runtimeObservedLaunch.executable_path) {
    summary.selected_launch_executable_path = runtimeObservedLaunch.executable_path;
  }
  summary.selected_launch_executable_ready = playwrightSelectedLaunchExecutableReady || runtimeObservedLaunch.executable_ready;
  if (playwrightSelectedLaunchExecutableBlockReason) {
    summary.selected_launch_executable_block_reason = playwrightSelectedLaunchExecutableBlockReason;
  }
  if (playwrightSelectedLaunchBlockReason) {
    summary.selected_launch_block_reason = playwrightSelectedLaunchBlockReason;
  }
  summary.delivery_transition_pending = playwrightDeliveryTransitionPending;
  if (playwrightDeliveryTransitionStage) {
    summary.delivery_transition_stage = playwrightDeliveryTransitionStage;
  }
  summary.launch_ready = playwrightLaunchReady;
  if (playwrightLaunchBlockReason) {
    summary.launch_block_reason = playwrightLaunchBlockReason;
  }
  summary.bundle_ready = playwrightBundleReady;
  summary.delivery_ready = playwrightDeliveryReady;
  if (playwrightCachePolicyVersion) {
    summary.policy_version = playwrightCachePolicyVersion;
  }
  if (playwrightCacheRetentionMode) {
    summary.retention_mode = playwrightCacheRetentionMode;
  }
  if (playwrightCacheRetainedDirs.length > 0) {
    summary.retained_dirs = playwrightCacheRetainedDirs;
  }
  if (playwrightCacheLastGCPrunedDirCount > 0) {
    summary.last_gc_pruned_dir_count = playwrightCacheLastGCPrunedDirCount;
  }
  if (playwrightBootstrapState) {
    summary.bootstrap_state = playwrightBootstrapState;
  }
  if (playwrightBootstrapErrorCode) {
    summary.bootstrap_error_code = playwrightBootstrapErrorCode;
  }
  summary.node_modules_ready = playwrightNodeModulesReady;
  summary.browser_ready = playwrightBrowserReady;
  const runtimeSummaryGeneration = playwrightRuntimeSummaryGeneration(summary);
  if (runtimeSummaryGeneration) {
    summary.runtime_summary_generation = runtimeSummaryGeneration;
  }
  const runtimeBaseline = playwrightRuntimeBaselineStatus(summary);
  summary.runtime_baseline_ready = runtimeBaseline.ready;
  if (runtimeBaseline.block_reason) {
    summary.runtime_baseline_block_reason = runtimeBaseline.block_reason;
  }
  return summary;
}

function runtimeObservedSelectedLaunchSummary() {
  const executablePath = runtimeObservedPlaywrightExecutablePath();
  if (!executablePath) {
    return {
      delivery_generation: '',
      source: '',
      ready: false,
      browser_revision: '',
      payload_source: '',
      payload_dirs: [],
      payload_ready: false,
      executable_path: '',
      executable_ready: false
    };
  }
  const browserRevision = playwrightChromiumRevisionForExecutablePath(playwrightBrowsersPath, executablePath);
  const deliveryGeneration = runtimeObservedSelectedLaunchDeliveryGeneration(browserRevision);
  return {
    delivery_generation: deliveryGeneration,
    source: 'runtime_observed',
    ready: true,
    browser_revision: browserRevision,
    payload_source: browserRevision ? 'active_browser_revision' : '',
    payload_dirs: runtimeObservedPlaywrightPayloadDirs(browserRevision),
    payload_ready: browserRevision !== '',
    executable_path: executablePath,
    executable_ready: true
  };
}

function runtimeObservedPlaywrightExecutablePath() {
  try {
    const executablePath = stringValue(runtimeState.playwright?.chromium?.executablePath?.());
    if (executablePath && existsSync(executablePath)) {
      return executablePath;
    }
  } catch {
    return '';
  }
  return '';
}

function runtimeObservedSelectedLaunchDeliveryGeneration(browserRevision) {
  const revision = stringValue(browserRevision);
  if (!revision || !playwrightBundleGeneration || !playwrightDependencyGeneration) {
    return '';
  }
  return browserdValueGenerationHash(playwrightBundleGeneration, playwrightDependencyGeneration, revision);
}

function playwrightChromiumRevisionForExecutablePath(cacheRoot, executablePath) {
  const normalizedRoot = stringValue(cacheRoot);
  const normalizedExecutable = stringValue(executablePath);
  if (!normalizedRoot || !normalizedExecutable) {
    return '';
  }
  try {
    const relative = path.relative(normalizedRoot, normalizedExecutable);
    if (!relative || relative === '.' || relative === '..' || relative.startsWith(`..${path.sep}`)) {
      return '';
    }
    const rootDir = stringValue(relative.split(path.sep)[0]);
    return playwrightChromiumRevisionSuffix(rootDir);
  } catch {
    return '';
  }
}

function runtimeObservedPlaywrightPayloadDirs(browserRevision) {
  const revision = stringValue(browserRevision);
  if (!revision || !playwrightBrowsersPath) {
    return [];
  }
  const candidates = [
    `chromium-${revision}`,
    `chromium_headless_shell-${revision}`
  ];
  return candidates.filter((candidate) => existsSync(path.join(playwrightBrowsersPath, candidate)));
}

function playwrightChromiumRevisionSuffix(dirName) {
  const normalized = stringValue(dirName);
  if (normalized.startsWith('chromium_headless_shell-')) {
    return stringValue(normalized.slice('chromium_headless_shell-'.length));
  }
  if (normalized.startsWith('chromium-')) {
    return stringValue(normalized.slice('chromium-'.length));
  }
  return '';
}

function browserdValueGenerationHash(...values) {
  const hash = createHash('sha256');
  for (const value of values) {
    hash.update(stringValue(value));
    hash.update('\0');
  }
  return hash.digest('hex').slice(0, 16);
}

function playwrightRuntimeSummaryGeneration(summary) {
  const cache = objectValue(summary);
  const hostOS = stringValue(cache.host_os);
  const hostArch = stringValue(cache.host_arch);
  const nodeVersion = stringValue(cache.node_version);
  const playwrightPackage = stringValue(cache.playwright_package);
  const playwrightPackageVersion = stringValue(cache.playwright_package_version);
  const selectedLaunchSource = stringValue(cache.selected_launch_source);
  const selectedLaunchDelivery = stringValue(cache.selected_launch_delivery_generation);
  if (!hostOS || !hostArch || !nodeVersion || !playwrightPackage || !playwrightPackageVersion || !selectedLaunchSource || !selectedLaunchDelivery) {
    return '';
  }
  return browserdValueGenerationHash(
    hostOS,
    hostArch,
    nodeVersion,
    playwrightPackage,
    playwrightPackageVersion,
    selectedLaunchSource,
    selectedLaunchDelivery,
    cache.selected_launch_executable_ready ? 'true' : 'false',
  );
}

function playwrightRuntimeBaselineStatus(summary) {
  const cache = objectValue(summary);
  if (!stringValue(cache.host_os) ||
      !stringValue(cache.host_arch) ||
      !stringValue(cache.node_version) ||
      !stringValue(cache.playwright_package) ||
      !stringValue(cache.playwright_package_version)) {
    return { ready: false, block_reason: 'runtime_host_summary_missing' };
  }
  if (!stringValue(cache.selected_launch_source)) {
    return { ready: false, block_reason: 'selected_launch_missing' };
  }
  if (!stringValue(cache.selected_launch_delivery_generation)) {
    return { ready: false, block_reason: 'selected_launch_delivery_missing' };
  }
  if (cache.selected_launch_payload_ready !== true) {
    return {
      ready: false,
      block_reason: stringValue(cache.selected_launch_payload_block_reason) || 'selected_launch_payload_not_ready'
    };
  }
  if (cache.selected_launch_executable_ready !== true) {
    return {
      ready: false,
      block_reason: stringValue(cache.selected_launch_executable_block_reason) || 'selected_launch_executable_not_ready'
    };
  }
  if (cache.selected_launch_ready !== true) {
    return {
      ready: false,
      block_reason: stringValue(cache.selected_launch_block_reason) || 'selected_launch_not_ready'
    };
  }
  if (!stringValue(cache.runtime_summary_generation)) {
    return { ready: false, block_reason: 'runtime_summary_generation_missing' };
  }
  return { ready: true, block_reason: '' };
}

function normalizedHostOS() {
  switch (process.platform) {
    case 'win32':
      return 'windows';
    default:
      return stringValue(process.platform);
  }
}

function normalizedHostArch() {
  switch (process.arch) {
    case 'x64':
      return 'amd64';
    default:
      return stringValue(process.arch);
  }
}

function normalizedNodeVersion() {
  return stringValue(process.versions?.node);
}

function resolvedPlaywrightPackageName() {
  return firstNonEmpty(runtimeState.playwrightPackage, installedPlaywrightPackageName());
}

function resolvedPlaywrightPackageVersion(packageName) {
  const resolvedName = firstNonEmpty(packageName, resolvedPlaywrightPackageName());
  if (!resolvedName) {
    return '';
  }
  try {
    const packagePath = path.join(scriptRoot, 'node_modules', resolvedName, 'package.json');
    if (!existsSync(packagePath)) {
      return '';
    }
    const payload = JSON.parse(readFileSync(packagePath, 'utf8'));
    return stringValue(payload?.version);
  } catch {
    return '';
  }
}

function installedPlaywrightPackageName() {
  for (const packageName of ['playwright', 'playwright-core']) {
    if (existsSync(path.join(scriptRoot, 'node_modules', packageName, 'package.json'))) {
      return packageName;
    }
  }
  return '';
}

function clearBrowserDisconnectStateForProfile(profileName) {
  const recoveredAt = Date.now();
  const profile = normalizedProfile(profileName);
  const retained = runtimeState.retainedSessionStateByProfile.get(profile);
  if (retained && typeof retained === 'object') {
    retained.browserDisconnect = null;
    const history = disconnectHistoryForState(retained);
    if (intValue(history.count) > 0) {
      retained.disconnectHistory = {
        ...history,
        lastRecoveredAt: recoveredAt
      };
    }
    if (Object.keys(objectValue(retained)).length > 0) {
      runtimeState.retainedSessionStateByProfile.set(profile, retained);
    }
  }
  const page = currentPageForProfile(profile);
  if (page) {
    const state = pageSessionStateForPage(page);
    state.browserDisconnect = null;
    const history = disconnectHistoryForState(state);
    if (intValue(history.count) > 0) {
      state.disconnectHistory = {
        ...history,
        lastRecoveredAt: recoveredAt
      };
    }
    state.updatedAt = recoveredAt;
  }
}

function beginDisconnectRestartAttempt(profileName) {
  const profile = normalizedProfile(profileName);
  const retained = cloneSessionState(currentSessionStateForProfile(profile)) || {};
  const history = disconnectHistoryForState(retained);
  if (intValue(history.count) <= 0) {
    return false;
  }
  const attemptedAt = Date.now();
  retained.disconnectHistory = {
    ...history,
    recentOccurredAt: normalizedDisconnectRecentOccurredAt(history.recentOccurredAt, attemptedAt),
    restartAttemptCount: intValue(history.restartAttemptCount) + 1,
    lastRestartAttemptAt: attemptedAt,
    lastRestartResult: 'pending',
    lastRestartError: ''
  };
  retained.updatedAt = attemptedAt;
  runtimeState.retainedSessionStateByProfile.set(profile, retained);
  return true;
}

function completeDisconnectRestartAttempt(profileName, result, errMessage = '') {
  const profile = normalizedProfile(profileName);
  const retained = cloneSessionState(currentSessionStateForProfile(profile)) || {};
  const history = disconnectHistoryForState(retained);
  if (intValue(history.restartAttemptCount) <= 0) {
    return;
  }
  const updatedAt = Date.now();
  retained.disconnectHistory = {
    ...history,
    recentOccurredAt: normalizedDisconnectRecentOccurredAt(history.recentOccurredAt, updatedAt),
    restartFailureCount: stringValue(result) === 'failed' ? restartFailureCountForHistory(history) + 1 : 0,
    lastRestartResult: stringValue(result),
    lastRestartError: stringValue(errMessage)
  };
  retained.updatedAt = updatedAt;
  runtimeState.retainedSessionStateByProfile.set(profile, retained);
}

function sessionDiagnosticsNote(page) {
  const tokens = [];
  appendSessionDiagnosticsTokens(tokens, page ? pageSessionStateForPage(page) : null);
  return joinNoteTokens(tokens);
}

function sessionHealthSummaryForPage(page) {
  return sessionHealthSummaryForState(page ? pageSessionStateForPage(page) : null, page);
}

function disconnectHistoryForState(state) {
  return objectValue(state?.disconnectHistory);
}

function normalizedDisconnectRecentOccurredAt(value, now = Date.now()) {
  const cutoff = now - disconnectBurstWindowMs;
  return arrayValue(value)
    .map((item) => intValue(item))
    .filter((item) => item > 0 && item >= cutoff)
    .sort((left, right) => left - right);
}

function disconnectBurstForHistory(history, now = Date.now()) {
  const recentOccurredAt = normalizedDisconnectRecentOccurredAt(history?.recentOccurredAt, now);
  return {
    count: recentOccurredAt.length,
    windowMs: disconnectBurstWindowMs,
    recentOccurredAt
  };
}

function disconnectBurstForState(state, now = Date.now()) {
  return disconnectBurstForHistory(disconnectHistoryForState(state), now);
}

function disconnectBackoffMsForCount(count) {
  const current = intValue(count);
  if (current <= 0) {
    return 0;
  }
  const exponent = Math.min(current-1, 5);
  return Math.min(1000 * (2 ** exponent), 30000);
}

function disconnectRecoveryActionForBurstCount(count) {
  return intValue(count) >= disconnectBurstThreshold ? 'browser action=restart' : 'browser action=ensure';
}

function disconnectReconnectHintForBurstCount(count) {
  return intValue(count) >= disconnectBurstThreshold ? 'restart_after_backoff' : 'ensure_after_backoff';
}

function disconnectCooldownRemainingMs(history, now = Date.now()) {
  const current = objectValue(history);
  const burst = disconnectBurstForHistory(current, now);
  if (intValue(burst.count) < disconnectBurstThreshold) {
    return 0;
  }
  const lastOccurredAt = intValue(current.lastOccurredAt);
  const recommendedBackoffMs = intValue(current.recommendedBackoffMs);
  if (lastOccurredAt <= 0 || recommendedBackoffMs <= 0) {
    return 0;
  }
  const remaining = (lastOccurredAt + recommendedBackoffMs) - now;
  if (remaining <= 0) {
    return 0;
  }
  return remaining;
}

function currentRestartFailure(state) {
  const current = objectValue(state);
  const history = disconnectHistoryForState(current);
  if (stringValue(history.lastRestartResult) !== 'failed') {
    return null;
  }
  if (intValue(history.lastRestartAttemptAt) < intValue(history.lastOccurredAt)) {
    return null;
  }
  return {
    message: stringValue(history.lastRestartError),
    attemptedAt: intValue(history.lastRestartAttemptAt)
  };
}

function restartFailureCountForHistory(history) {
  return intValue(objectValue(history).restartFailureCount);
}

function restartFailureEscalated(history) {
  const current = objectValue(history);
  if (stringValue(current.lastRestartResult) !== 'failed') {
    return false;
  }
  return restartFailureCountForHistory(current) >= restartFailurePermanentThreshold;
}

function restartRetryBackoffRemainingMs(history, now = Date.now()) {
  const current = objectValue(history);
  if (stringValue(current.lastRestartResult) !== 'failed') {
    return 0;
  }
  const lastRestartAttemptAt = intValue(current.lastRestartAttemptAt);
  const recommendedBackoffMs = intValue(current.recommendedBackoffMs);
  if (lastRestartAttemptAt <= 0 || recommendedBackoffMs <= 0) {
    return 0;
  }
  const remaining = (lastRestartAttemptAt + recommendedBackoffMs) - now;
  if (remaining <= 0) {
    return 0;
  }
  return remaining;
}

function enforceDisconnectCooldown(profileName) {
  const profile = normalizedProfile(profileName);
  const history = disconnectHistoryForState(currentSessionStateForProfile(profile));
  const remaining = disconnectCooldownRemainingMs(history);
  if (remaining <= 0) {
    return;
  }
  throw httpError(429, `browser restart cooldown active; retry after ${remaining}ms`);
}

function enforceRestartRetryBackoff(profileName) {
  const profile = normalizedProfile(profileName);
  const history = disconnectHistoryForState(currentSessionStateForProfile(profile));
  const remaining = restartRetryBackoffRemainingMs(history);
  if (remaining <= 0) {
    return;
  }
  throw httpError(429, `browser restart retry backoff active; retry after ${remaining}ms`);
}

function enforcePermanentRestartFailure(profileName) {
  const profile = normalizedProfile(profileName);
  const history = disconnectHistoryForState(currentSessionStateForProfile(profile));
  if (!restartFailureEscalated(history)) {
    return;
  }
  throw httpError(409, 'browser restart permanently failed; explicit browser.start required');
}

function restartFailureHookPath() {
  return path.join(stateRoot, 'test-hooks', 'fail-next-restart.txt');
}

async function maybeInjectRestartFailure(profileName) {
  if (!enableTestHooks) {
    return;
  }
  const profile = normalizedProfile(profileName);
  const hookPath = restartFailureHookPath();
  let raw = '';
  try {
    raw = await fs.readFile(hookPath, 'utf8');
  } catch (err) {
    if (err && err.code === 'ENOENT') {
      return;
    }
    throw err;
  }
  await fs.rm(hookPath, { force: true });
  const configuredMessage = stringValue(raw);
  throw httpError(503, configuredMessage || `injected restart failure for profile ${profile}`);
}

function sessionHealthSummaryForState(state, page) {
  if (!state || typeof state !== 'object') {
    return null;
  }
  const summary = {};
  const assessment = sessionHealthAssessment(state, page);
  const disconnectHistory = disconnectHistoryForState(state);
  const disconnectBurst = disconnectBurstForState(state);
  const cooldownRemainingMs = disconnectCooldownRemainingMs(disconnectHistory);
  const permanentRestartFailure = restartFailureEscalated(disconnectHistory);
  const retryBackoffRemainingMs = permanentRestartFailure ? 0 : restartRetryBackoffRemainingMs(disconnectHistory);
  if (stringValue(assessment.state)) {
    summary.state = stringValue(assessment.state);
  }
  if (stringValue(assessment.reason)) {
    summary.reason = stringValue(assessment.reason);
  }
  if (stringValue(assessment.recoveryAction)) {
    summary.recovery_action = stringValue(assessment.recoveryAction);
  }
  if (stringValue(assessment.reconnectHint)) {
    summary.reconnect_hint = stringValue(assessment.reconnectHint);
  } else if (intValue(disconnectBurst.count) >= disconnectBurstThreshold) {
    summary.reconnect_hint = disconnectReconnectHintForBurstCount(disconnectBurst.count);
  }
  if (intValue(disconnectHistory.count) > 0) {
    summary.disconnect_count = intValue(disconnectHistory.count);
  }
  if (intValue(disconnectBurst.count) >= disconnectBurstThreshold) {
    summary.disconnect_burst_count = intValue(disconnectBurst.count);
    summary.disconnect_burst_window_ms = intValue(disconnectBurst.windowMs);
  }
  if (cooldownRemainingMs > 0) {
    summary.cooldown_remaining_ms = cooldownRemainingMs;
  }
  if (retryBackoffRemainingMs > 0) {
    summary.retry_backoff_remaining_ms = retryBackoffRemainingMs;
  }
  if (intValue(disconnectHistory.restartAttemptCount) > 0) {
    summary.restart_attempt_count = intValue(disconnectHistory.restartAttemptCount);
  }
  if (restartFailureCountForHistory(disconnectHistory) > 0) {
    summary.restart_failure_count = restartFailureCountForHistory(disconnectHistory);
  }
  if (intValue(disconnectHistory.lastOccurredAt) > 0) {
    summary.last_disconnect_unix_milli = intValue(disconnectHistory.lastOccurredAt);
  }
  if (intValue(disconnectHistory.lastRecoveredAt) > 0) {
    summary.last_reconnect_unix_milli = intValue(disconnectHistory.lastRecoveredAt);
  }
  if (intValue(disconnectHistory.lastRestartAttemptAt) > 0) {
    summary.last_restart_attempt_unix_milli = intValue(disconnectHistory.lastRestartAttemptAt);
  }
  if (stringValue(disconnectHistory.lastRestartResult)) {
    summary.last_restart_result = stringValue(disconnectHistory.lastRestartResult);
  }
  if (stringValue(disconnectHistory.lastRestartError)) {
    summary.last_restart_error = stringValue(disconnectHistory.lastRestartError);
  }
  if (intValue(disconnectHistory.recommendedBackoffMs) > 0) {
    summary.recommended_backoff_ms = intValue(disconnectHistory.recommendedBackoffMs);
  }
  const events = arrayValue(state.events);
  if (events.length > 0) {
    summary.events_buffered = events.length;
  }
  if (stringValue(state.lastLifecycleEvent)) {
    summary.last_event = stringValue(state.lastLifecycleEvent);
  }
  if (intValue(state.popupCount) > 0) {
    summary.popup_count = intValue(state.popupCount);
  }
  if (state.lastDialog && typeof state.lastDialog === 'object') {
    if (stringValue(state.lastDialog.action)) {
      summary.last_dialog_action = stringValue(state.lastDialog.action);
    }
    if (stringValue(state.lastDialog.type)) {
      summary.last_dialog_type = stringValue(state.lastDialog.type);
    }
    if (stringValue(state.lastDialog.source)) {
      summary.last_dialog_source = stringValue(state.lastDialog.source);
    }
  }
  if (state.lastDownload && typeof state.lastDownload === 'object') {
    if (stringValue(state.lastDownload.suggestedFilename)) {
      summary.last_download_filename = stringValue(state.lastDownload.suggestedFilename);
    }
    if (stringValue(state.lastDownload.outputMode)) {
      summary.last_download_output = stringValue(state.lastDownload.outputMode);
    }
  }
  if (page && !page.isClosed()) {
    summary.current_target_id = targetIDForPage(page);
    const tabIndex = currentTabIndexForPage(page);
    if (tabIndex > 0) {
      summary.current_tab_index = tabIndex;
    }
  }
  const pendingDownloads = arrayValue(state.pendingDownloads).length;
  if (pendingDownloads > 0) {
    summary.pending_downloads = pendingDownloads;
  }
  if (state.pendingDialog && typeof state.pendingDialog === 'object' && stringValue(state.pendingDialog.action)) {
    summary.pending_dialog_action = stringValue(state.pendingDialog.action);
  }
  if (state.staleTarget && typeof state.staleTarget === 'object') {
    if (stringValue(state.staleTarget.status)) {
      summary.stale_target_resolver_status = stringValue(state.staleTarget.status);
    }
    if (stringValue(state.staleTarget.blockedBy)) {
      summary.stale_target_blocked_by = stringValue(state.staleTarget.blockedBy);
    }
  }
  if (intValue(state.updatedAt) > 0) {
    summary.last_updated_unix_milli = intValue(state.updatedAt);
  }
  return Object.keys(summary).length > 0 ? summary : null;
}

function sessionHealthAssessment(state, page) {
  const browserDisconnect = objectValue(state?.browserDisconnect);
  const disconnectHistory = disconnectHistoryForState(state);
  const disconnectBurst = disconnectBurstForState(state);
  const restartFailure = currentRestartFailure(state);
  const cooldownRemainingMs = disconnectCooldownRemainingMs(disconnectHistory);
  const permanentRestartFailure = restartFailureEscalated(disconnectHistory);
  const retryBackoffRemainingMs = permanentRestartFailure ? 0 : restartRetryBackoffRemainingMs(disconnectHistory);
  const popupCount = intValue(state?.popupCount);
  const pendingDialogAction = stringValue(state?.pendingDialog?.action);
  const lastEvent = stringValue(state?.lastLifecycleEvent);
  const staleTarget = objectValue(state?.staleTarget);
  const latestPageError = latestSessionEventByCategory(state, 'script');
  if (stringValue(browserDisconnect.message)) {
    if (cooldownRemainingMs > 0) {
      return {
        state: 'cooldown_active',
        reason: `browser restart cooldown active for ${cooldownRemainingMs}ms after ${intValue(disconnectBurst.count)} disconnects`,
        recoveryAction: 'browser action=wait',
        reconnectHint: 'retry_after_cooldown'
      };
    }
    if (permanentRestartFailure && restartFailure && stringValue(restartFailure.message)) {
      return {
        state: 'restart_failed_permanent',
        reason: `browser restart failed ${restartFailureCountForHistory(disconnectHistory)} times; explicit browser.start required`,
        recoveryAction: 'browser action=start',
        reconnectHint: 'manual_restart_required'
      };
    }
    if (restartFailure && stringValue(restartFailure.message) && retryBackoffRemainingMs > 0) {
      return {
        state: 'restart_pending',
        reason: `browser restart retry pending for ${retryBackoffRemainingMs}ms after failure`,
        recoveryAction: 'browser action=wait',
        reconnectHint: 'retry_after_backoff'
      };
    }
    if (restartFailure && stringValue(restartFailure.message)) {
      return {
        state: 'restart_failed',
        reason: truncate(restartFailure.message, 120),
        recoveryAction: disconnectRecoveryActionForBurstCount(disconnectBurst.count || disconnectHistory.count || 1),
        reconnectHint: disconnectReconnectHintForBurstCount(disconnectBurst.count || disconnectHistory.count || 1)
      };
    }
    if (intValue(disconnectBurst.count) >= disconnectBurstThreshold) {
      return {
        state: 'browser_disconnect_burst',
        reason: `browser runtime disconnected ${intValue(disconnectBurst.count)} times within ${intValue(disconnectBurst.windowMs)}ms`,
        recoveryAction: disconnectRecoveryActionForBurstCount(disconnectBurst.count),
        reconnectHint: disconnectReconnectHintForBurstCount(disconnectBurst.count)
      };
    }
    return {
      state: 'browser_disconnected',
      reason: truncate(stringValue(browserDisconnect.message), 120),
      recoveryAction: stringValue(browserDisconnect.recoveryAction) || 'browser action=ensure',
      reconnectHint: stringValue(browserDisconnect.reconnectHint) || disconnectReconnectHintForBurstCount(1)
    };
  }
  if (pendingDialogAction) {
    return {
      state: 'dialog_pending',
      reason: `browser page is waiting for dialog action ${pendingDialogAction}`,
      recoveryAction: 'browser action=dialog'
    };
  }
  if (popupCount >= 2) {
    return {
      state: 'popup_pressure',
      reason: `browser page opened ${popupCount} popup targets`,
      recoveryAction: 'browser action=tabs'
    };
  }
  if (lastEvent === 'closed') {
    return {
      state: 'page_closed',
      reason: 'browser page closed',
      recoveryAction: 'browser action=ensure'
    };
  }
  if (stringValue(staleTarget.message)) {
    return {
      state: 'stale_target',
      reason: truncate(stringValue(staleTarget.message), 120),
      recoveryAction: stringValue(staleTarget.recoveryAction) || 'browser action=snapshot'
    };
  }
  if (latestPageError && stringValue(latestPageError.message)) {
    return {
      state: 'page_error',
      reason: truncate(stringValue(latestPageError.message), 120),
      recoveryAction: 'browser action=refresh'
    };
  }
  if (page && !page.isClosed()) {
    return {
      state: 'healthy',
      reason: 'browser page is available',
      recoveryAction: ''
    };
  }
  return {
    state: '',
    reason: '',
    recoveryAction: ''
  };
}

function latestSessionEventByCategory(state, category) {
  const targetCategory = stringValue(category);
  if (!targetCategory) {
    return null;
  }
  const events = arrayValue(state?.events);
  for (let idx = events.length - 1; idx >= 0; idx -= 1) {
    const current = objectValue(events[idx]);
    if (stringValue(current.category) === targetCategory) {
      return current;
    }
  }
  return null;
}

function appendSessionDiagnosticsTokens(target, state) {
  if (!state || typeof state !== 'object') {
    return;
  }
  const disconnectHistory = disconnectHistoryForState(state);
  const disconnectBurst = disconnectBurstForState(state);
  const restartFailure = currentRestartFailure(state);
  const cooldownRemainingMs = disconnectCooldownRemainingMs(disconnectHistory);
  const permanentRestartFailure = restartFailureEscalated(disconnectHistory);
  const retryBackoffRemainingMs = permanentRestartFailure ? 0 : restartRetryBackoffRemainingMs(disconnectHistory);
  if (state.browserDisconnect && typeof state.browserDisconnect === 'object' && stringValue(state.browserDisconnect.message)) {
    target.push('browser_disconnected=1');
  }
  if (permanentRestartFailure && restartFailure && stringValue(restartFailure.message)) {
    target.push('restart_failed_permanent=1');
  }
  if (restartFailure && stringValue(restartFailure.message) && retryBackoffRemainingMs > 0) {
    target.push('restart_pending=1');
  }
  if (restartFailure && stringValue(restartFailure.message)) {
    target.push('restart_failed=1');
  }
  if (intValue(disconnectHistory.count) > 0) {
    target.push(noteToken('disconnect_count', String(intValue(disconnectHistory.count))));
  }
  if (intValue(disconnectBurst.count) >= disconnectBurstThreshold) {
    target.push(noteToken('disconnect_burst_count', String(intValue(disconnectBurst.count))));
    target.push(noteToken('disconnect_burst_window_ms', String(intValue(disconnectBurst.windowMs))));
    target.push(noteToken('reconnect_hint', permanentRestartFailure ? 'manual_restart_required' : (cooldownRemainingMs > 0 ? 'retry_after_cooldown' : (retryBackoffRemainingMs > 0 ? 'retry_after_backoff' : disconnectReconnectHintForBurstCount(disconnectBurst.count)))));
  }
  if (cooldownRemainingMs > 0) {
    target.push(noteToken('cooldown_remaining_ms', String(cooldownRemainingMs)));
  }
  if (retryBackoffRemainingMs > 0) {
    target.push(noteToken('retry_backoff_remaining_ms', String(retryBackoffRemainingMs)));
  }
  if (intValue(disconnectHistory.lastOccurredAt) > 0) {
    target.push(noteToken('last_disconnect_unix_milli', String(intValue(disconnectHistory.lastOccurredAt))));
  }
  if (intValue(disconnectHistory.lastRecoveredAt) > 0) {
    target.push(noteToken('last_reconnect_unix_milli', String(intValue(disconnectHistory.lastRecoveredAt))));
  }
  if (intValue(disconnectHistory.restartAttemptCount) > 0) {
    target.push(noteToken('restart_attempt_count', String(intValue(disconnectHistory.restartAttemptCount))));
  }
  if (restartFailureCountForHistory(disconnectHistory) > 0) {
    target.push(noteToken('restart_failure_count', String(restartFailureCountForHistory(disconnectHistory))));
  }
  if (intValue(disconnectHistory.lastRestartAttemptAt) > 0) {
    target.push(noteToken('last_restart_attempt_unix_milli', String(intValue(disconnectHistory.lastRestartAttemptAt))));
  }
  if (stringValue(disconnectHistory.lastRestartResult)) {
    target.push(noteToken('restart_result', stringValue(disconnectHistory.lastRestartResult)));
  }
  if (stringValue(disconnectHistory.lastRestartError)) {
    target.push(noteToken('last_restart_error', stringValue(disconnectHistory.lastRestartError)));
  }
  if (intValue(disconnectHistory.recommendedBackoffMs) > 0) {
    target.push(noteToken('recommended_backoff_ms', String(intValue(disconnectHistory.recommendedBackoffMs))));
  }
  const events = arrayValue(state.events);
  if (events.length > 0) {
    target.push(`events_buffered=${events.length}`);
  }
  if (stringValue(state.lastLifecycleEvent)) {
    target.push(noteToken('last_event', state.lastLifecycleEvent));
  }
  if (intValue(state.popupCount) > 0) {
    target.push(`popup_count=${intValue(state.popupCount)}`);
  }
  if (state.lastDialog && typeof state.lastDialog === 'object') {
    if (stringValue(state.lastDialog.action)) {
      target.push(noteToken('last_dialog_action', state.lastDialog.action));
    }
    if (stringValue(state.lastDialog.type)) {
      target.push(noteToken('last_dialog_type', state.lastDialog.type));
    }
    if (stringValue(state.lastDialog.source)) {
      target.push(noteToken('last_dialog_source', state.lastDialog.source));
    }
  }
  if (state.lastDownload && typeof state.lastDownload === 'object' && stringValue(state.lastDownload.suggestedFilename)) {
    target.push(noteToken('last_download_filename', state.lastDownload.suggestedFilename));
    if (stringValue(state.lastDownload.outputMode)) {
      target.push(noteToken('last_download_output', state.lastDownload.outputMode));
    }
  }
  if (state.staleTarget && typeof state.staleTarget === 'object') {
    if (stringValue(state.staleTarget.status)) {
      target.push(noteToken('stale_target_resolver_status', state.staleTarget.status));
    }
    if (stringValue(state.staleTarget.blockedBy)) {
      target.push(noteToken('stale_target_blocked_by', state.staleTarget.blockedBy));
    }
  }
}

function dialogResultNoteTokens(action, promptText) {
  const meta = dialogResultMetadata(action, promptText);
  const tokens = [];
  if (stringValue(meta?.action)) {
    tokens.push(noteToken('dialog_action', meta.action));
  }
  if (stringValue(meta?.prompt_state)) {
    tokens.push(noteToken('dialog_prompt', meta.prompt_state));
  }
  return tokens;
}

function downloadResultNoteTokens(mode, record) {
  const meta = downloadResultMetadata(mode, record);
  const tokens = [];
  if (stringValue(meta?.mode)) {
    tokens.push(noteToken('download_mode', meta.mode));
  }
  if (stringValue(meta?.suggested_filename)) {
    tokens.push(noteToken('download_filename', meta.suggested_filename));
  }
  if (stringValue(meta?.output_mode)) {
    tokens.push(noteToken('download_output', meta.output_mode));
  }
  return tokens;
}

function uploadResultNoteTokens(mode, count) {
  const tokens = [];
  if (stringValue(mode)) {
    tokens.push(noteToken('upload_mode', mode));
  }
  if (count > 0) {
    tokens.push(noteToken('upload_count', String(count)));
  }
  return tokens;
}

function dialogResultMetadata(action, promptText) {
  const meta = {};
  if (stringValue(action)) {
    meta.action = stringValue(action);
  }
  if (stringValue(action) === 'accept') {
    meta.prompt_state = stringValue(promptText) ? 'provided' : 'empty';
  }
  return Object.keys(meta).length > 0 ? meta : null;
}

function downloadResultMetadata(mode, record) {
  const meta = {};
  if (stringValue(mode)) {
    meta.mode = stringValue(mode);
  }
  if (stringValue(record?.suggestedFilename)) {
    meta.suggested_filename = stringValue(record.suggestedFilename);
  }
  if (stringValue(record?.outputMode)) {
    meta.output_mode = stringValue(record.outputMode);
  }
  if (stringValue(record?.artifactPath)) {
    meta.backend_path = stringValue(record.artifactPath);
  }
  if (intValue(record?.byteSize) > 0) {
    meta.byte_size = intValue(record.byteSize);
  }
  return Object.keys(meta).length > 0 ? meta : null;
}

function joinNoteTokens(tokens) {
  return arrayValue(tokens)
    .map((item) => stringValue(item))
    .filter(Boolean)
    .join(' ');
}

function noteToken(key, value) {
  const normalizedKey = stringValue(key).replace(/[^a-z0-9_]+/gi, '_').toLowerCase();
  const normalizedValue = noteTokenValue(value);
  if (!normalizedKey || !normalizedValue) {
    return '';
  }
  return `${normalizedKey}=${normalizedValue}`;
}

function noteTokenValue(value) {
  const current = stringValue(value);
  if (!current) {
    return '';
  }
  return current.replace(/[^a-zA-Z0-9._-]+/g, '_').replace(/^_+|_+$/g, '');
}

function browserNavigateDownloadStarted(err) {
  const message = stringValue(err && err.message ? err.message : err).toLowerCase();
  return message.includes('download is starting');
}

function browserNavigateShouldFallbackToDomContentLoaded(err) {
  const message = stringValue(err && err.message ? err.message : err).toLowerCase();
  return message.includes('page.goto: timeout') && message.includes('waiting until "load"');
}

async function gotoPageWithLoadFallback(page, url, timeout) {
  try {
    await page.goto(url, {
      waitUntil: 'load',
      timeout
    });
    return {
      downloadStarted: false,
      waitFallback: ''
    };
  } catch (err) {
    if (browserNavigateDownloadStarted(err)) {
      return {
        downloadStarted: true,
        waitFallback: ''
      };
    }
    if (!browserNavigateShouldFallbackToDomContentLoaded(err)) {
      throw err;
    }
  }
  await page.goto(url, {
    waitUntil: 'domcontentloaded',
    timeout
  });
  return {
    downloadStarted: false,
    waitFallback: 'domcontentloaded'
  };
}

function normalizeDialogAction(action) {
  const current = stringValue(action).toLowerCase();
  if (current === 'accept' || current === 'dismiss') {
    return current;
  }
  return '';
}

async function handlePageDialog(page, dialog) {
  const state = pageSessionStateForPage(page);
  const armed = state.pendingDialog && typeof state.pendingDialog === 'object' ? state.pendingDialog : null;
  state.pendingDialog = null;
  runtimeState.activePage = page;
  const action = normalizeDialogAction(armed?.action) || 'dismiss';
  const promptText = stringValue(armed?.promptText);
  state.lastDialog = {
    action,
    type: safeCall(() => dialog.type()),
    message: safeCall(() => dialog.message()),
    defaultValue: safeCall(() => dialog.defaultValue()),
    source: armed ? 'armed' : 'auto_dismissed',
    at: Date.now()
  };
  state.updatedAt = state.lastDialog.at;
  try {
    if (action === 'accept') {
      if (promptText) {
        await dialog.accept(promptText);
      } else {
        await dialog.accept();
      }
      return;
    }
    await dialog.dismiss();
  } catch {
    return;
  }
}

async function handlePageFileChooser(page, chooser) {
  const state = pageSessionStateForPage(page);
  const armed = state.pendingUpload && typeof state.pendingUpload === 'object' ? state.pendingUpload : null;
  state.pendingUpload = null;
  runtimeState.activePage = page;
  const paths = arrayValue(armed?.paths).map((item) => stringValue(item)).filter(Boolean);
  state.updatedAt = Date.now();
  if (paths.length === 0) {
    return;
  }
  try {
    await chooser.setFiles(paths);
  } catch {
    return;
  }
}

function handlePageDownload(page, download) {
  const state = pageSessionStateForPage(page);
  const seq = state.nextDownloadSeq + 1;
  state.nextDownloadSeq = seq;
  const entry = {
    seq,
    promise: captureDownloadRecord(page, download)
  };
  state.pendingDownloads.push(entry);
  recordPageLifecycle(page, 'download_started');
  entry.promise.then((record) => {
    state.lastDownload = record;
    state.updatedAt = Date.now();
  }).catch(() => {
    state.updatedAt = Date.now();
  });
  resolvePendingDownloadWaiters(state);
}

async function captureDownloadRecord(page, download) {
  const suggestedFilename = stringValue(safeCall(() => download.suggestedFilename())) || 'download.bin';
  const downloadURL = stringValue(safeCall(() => download.url()));
  const targetPath = artifactPath('download', downloadArtifactExtension(suggestedFilename, downloadURL));
  await fs.mkdir(path.dirname(targetPath), { recursive: true });
  await download.saveAs(targetPath);
  const info = await fs.stat(targetPath);
  return {
    artifactPath: targetPath,
    path: targetPath,
    finalURL: firstNonEmpty(downloadURL, stringValue(safeCall(() => page.url()))),
    title: await safeAsyncString(async () => await page.title()),
    contentType: contentTypeFromHints(suggestedFilename, downloadURL, targetPath),
    suggestedFilename,
    outputMode: 'artifact',
    byteSize: info && typeof info.size === 'number' ? info.size : 0,
    recordedAt: Date.now()
  };
}

function waitForDownloadEntry(page, minSeq, timeout, options = {}) {
  const state = pageSessionStateForPage(page);
  const queued = takeDownloadEntry(state, minSeq);
  if (queued) {
    return Promise.resolve({ entry: queued, mode: 'wait' });
  }
  const recent = takeRecentDownloadEntry(state, options);
  if (recent) {
    return Promise.resolve({ entry: recent, mode: 'wait_recent_reuse' });
  }
  return new Promise((resolve, reject) => {
    const waiter = {
      minSeq,
      resolve,
      reject,
      allowRecentDownloadReuse: options && options.allowRecentDownloadReuse === true,
      timer: null
    };
    if (timeout > 0) {
      waiter.timer = setTimeout(() => {
        removeDownloadWaiter(state, waiter);
        const recent = takeRecentDownloadEntry(state, { allowRecentDownloadReuse: waiter.allowRecentDownloadReuse });
        if (recent) {
          waiter.resolve({ entry: recent, mode: 'wait_recent_reuse' });
          return;
        }
        reject(httpError(408, `download did not start within ${timeout}ms`));
      }, timeout);
    }
    state.downloadWaiters.push(waiter);
  });
}

function takeDownloadEntry(state, minSeq) {
  const index = state.pendingDownloads.findIndex((entry) => intValue(entry?.seq) > minSeq);
  if (index < 0) {
    return null;
  }
  const [entry] = state.pendingDownloads.splice(index, 1);
  return entry || null;
}

function resolvePendingDownloadWaiters(state) {
  if (!Array.isArray(state.downloadWaiters) || state.downloadWaiters.length === 0) {
    return;
  }
  const remaining = [];
  for (const waiter of state.downloadWaiters) {
    const entry = takeDownloadEntry(state, waiter.minSeq);
    if (!entry) {
      remaining.push(waiter);
      continue;
    }
    if (waiter.timer) {
      clearTimeout(waiter.timer);
    }
    waiter.resolve({ entry, mode: 'wait' });
  }
  state.downloadWaiters = remaining;
}

function removeDownloadWaiter(state, waiter) {
  state.downloadWaiters = arrayValue(state.downloadWaiters).filter((current) => current !== waiter);
}

function takeRecentDownloadEntry(state, options = {}) {
  if (!state || options.allowRecentDownloadReuse !== true) {
    return null;
  }
  const localEntry = reusableRecentDownloadEntryForState(state);
  if (localEntry) {
    return localEntry;
  }
  return reusableRecentDownloadEntryForRuntime(state);
}

function markDownloadRecordWaitConsumed(record, mode) {
  const next = {
    ...objectValue(record),
    waitConsumedAt: Date.now()
  };
  const normalizedMode = stringValue(mode);
  if (normalizedMode) {
    next.waitMode = normalizedMode;
  }
  return next;
}

function reusableRecentDownloadEntryForState(state) {
  const reusable = normalizeReusableDownloadRecord(objectValue(state?.lastDownload));
  if (!reusable) {
    return null;
  }
  return {
    seq: intValue(reusable.seq),
    promise: Promise.resolve(reusable)
  };
}

function reusableRecentDownloadEntryForRuntime(excludedState) {
  let best = null;
  for (const state of runtimeState.pageSessionState.values()) {
    if (state === excludedState) {
      continue;
    }
    const current = normalizeReusableDownloadRecord(objectValue(state?.lastDownload));
    if (!current) {
      continue;
    }
    if (!best || intValue(current.recordedAt) > intValue(best.recordedAt)) {
      best = current;
    }
  }
  if (!best) {
    const retained = normalizeReusableDownloadRecord(objectValue(retainedSessionStateForProfile(runtimeState.activeProfile)?.lastDownload));
    if (retained) {
      best = retained;
    }
  }
  if (!best) {
    return null;
  }
  return {
    seq: intValue(best.seq),
    promise: Promise.resolve(best)
  };
}

function normalizeReusableDownloadRecord(record) {
  if (!record || typeof record !== 'object') {
    return null;
  }
  if (!stringValue(record.path) && !stringValue(record.artifactPath)) {
    return null;
  }
  const recordedAt = intValue(record.recordedAt);
  if (recordedAt <= 0 || Date.now() - recordedAt > recentDownloadReuseWindowMs) {
    return null;
  }
  const reusable = {
    ...record,
    path: firstNonEmpty(stringValue(record.artifactPath), stringValue(record.path))
  };
  if (!stringValue(reusable.path)) {
    return null;
  }
  return reusable;
}

async function finalizeDownloadEntry(entry, requestedPath) {
  const record = await entry.promise;
  const outputPath = stringValue(requestedPath);
  if (!outputPath || outputPath === record.path) {
    return record;
  }
  await fs.mkdir(path.dirname(outputPath), { recursive: true });
  await fs.copyFile(record.path, outputPath);
  const info = await fs.stat(outputPath);
  return {
    ...record,
    path: outputPath,
    outputMode: 'requested_path',
    byteSize: info && typeof info.size === 'number' ? info.size : record.byteSize
  };
}

function downloadArtifactExtension(...hints) {
  for (const item of hints) {
    const candidate = stringValue(item);
    if (!candidate) {
      continue;
    }
    const ext = path.extname(urlPath(candidate) || candidate);
    if (validArtifactExtension(ext)) {
      return ext;
    }
  }
  return '.bin';
}

function validArtifactExtension(value) {
  const current = stringValue(value).toLowerCase();
  if (!current || current === '.' || current.length > 12 || !current.startsWith('.')) {
    return false;
  }
  for (const ch of current.slice(1)) {
    if ((ch < 'a' || ch > 'z') && (ch < '0' || ch > '9')) {
      return false;
    }
  }
  return true;
}

function contentTypeFromHints(...hints) {
  const ext = downloadArtifactExtension(...hints);
  switch (ext) {
    case '.zip':
      return 'application/zip';
    case '.pdf':
      return 'application/pdf';
    case '.json':
      return 'application/json';
    case '.csv':
      return 'text/csv';
    case '.txt':
      return 'text/plain';
    case '.html':
    case '.htm':
      return 'text/html';
    case '.png':
      return 'image/png';
    case '.jpg':
    case '.jpeg':
      return 'image/jpeg';
    default:
      return 'application/octet-stream';
  }
}

function nativeRefStoreForPage(page) {
  const targetID = targetIDForPage(page);
  let refs = runtimeState.pageNativeRefs.get(targetID);
  if (!refs) {
    refs = new Map();
    runtimeState.pageNativeRefs.set(targetID, refs);
  }
  return refs;
}

function registerSnapshotNativeRef(page, element, pageBinding) {
  const entry = snapshotNativeRefEntry(element, pageBinding);
  if (!entry) {
    return '';
  }
  const ref = `e${runtimeState.nextNativeRefID}`;
  runtimeState.nextNativeRefID += 1;
  nativeRefStoreForPage(page).set(ref, entry);
  return ref;
}

function snapshotNativeRefEntry(element, pageBinding) {
  const selector = stringValue(element?.Selector || element?.selector);
  const selectorIndex = intValue(element?.SelectorIndex ?? element?.selector_index);
  const role = stringValue(element?.Role || element?.role);
  const tag = stringValue(element?.Tag || element?.tag);
  const label = stringValue(element?.Label || element?.label);
  const type = stringValue(element?.Type || element?.type);
  const href = stringValue(element?.Href || element?.href);
  const placeholder = stringValue(element?.Placeholder || element?.placeholder);
  const framePath = stringValue(element?.FramePath || element?.frame_path);
  const binding = normalizePageBindingCandidate(pageBinding);
  if (!selector && !tag && !label && !type && !href && !placeholder && !framePath && !binding) {
    return null;
  }
  return {
    selector,
    selector_index: selectorIndex > 0 ? selectorIndex : 0,
    frame_path: framePath,
    role,
    tag,
    label,
    type,
    href,
    placeholder,
    page_url: stringValue(binding?.page_url),
    page_origin: stringValue(binding?.page_origin),
    page_path: stringValue(binding?.page_path),
    page_title: stringValue(binding?.page_title),
    tab_index: intValue(binding?.tab_index)
  };
}

function nativeRefEntryForPage(page, nativeRef) {
  const targetID = pageTargetIDs.get(page);
  if (!targetID) {
    return null;
  }
  const refs = runtimeState.pageNativeRefs.get(targetID);
  if (!refs) {
    return null;
  }
  return refs.get(nativeRef) || null;
}

function locatorForNativeRef(page, nativeRef) {
  const entry = nativeRefEntryForPage(page, nativeRef);
  if (!entry) {
    return { locator: null, outcome: null };
  }
  const scope = scopeForNativeRefEntry(page, entry);
  if (!scope) {
    return {
      locator: null,
      outcome: normalizeResolverOutcome({
        status: 'page_binding_blocked',
        blocked_by: 'frame_path',
        note: `element ref expects frame path ${stringValue(entry.frame_path) || 'unknown'} but current frame tree no longer matches`
      })
    };
  }
  const candidate = nativeRefPrimaryLocatorCandidate(entry);
  if (candidate) {
    return {
      locator: locatorForCandidateInScope(scope, candidate),
      outcome: null,
      frame_path: stringValue(entry.frame_path),
      candidate,
      fallback_candidates: nativeRefFallbackLocatorCandidates(entry, candidate),
      scope
    };
  }
  return { locator: null, outcome: null };
}

function nativeRefFallbackLocatorCandidates(entry, primaryCandidate) {
  const primary = normalizeLocatorCandidate(primaryCandidate);
  if (!locatorCandidateValid(primary)) {
    return [];
  }
  const out = [];
  const variants = [primary];
  for (const relaxer of nativeRefFallbackRelaxersForCandidate(primary)) {
    const snapshot = variants.slice();
    for (const candidate of snapshot) {
      const relaxed = relaxer(candidate);
      const normalized = normalizeLocatorCandidate(relaxed);
      if (!locatorCandidateValid(normalized)) {
        continue;
      }
      const key = locatorCandidateKey(normalized);
      if (!key) {
        continue;
      }
      if (variants.some((item) => locatorCandidateKey(item) === key)) {
        continue;
      }
      variants.push(normalized);
      out.push(normalized);
    }
  }
  return out;
}

function nativeRefFallbackRelaxersForCandidate(primaryCandidate) {
  switch (stringValue(primaryCandidate?.kind)) {
    case 'role_label':
    case 'tag_label':
    case 'label':
      return [
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'href'),
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'placeholder'),
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'type')
      ];
    case 'placeholder':
      return [
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'type'),
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'placeholder')
      ];
    case 'tag_type':
      return [
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'type')
      ];
    case 'type':
      return [
        (candidate) => locatorCandidateRelaxedByClearingStringField(candidate, 'type')
      ];
    default:
      return [];
  }
}

function locatorCandidateRelaxedByClearingStringField(candidate, field) {
  const current = normalizeLocatorCandidate(candidate);
  if (!locatorCandidateValid(current) || !stringValue(current[field])) {
    return null;
  }
  const next = {
    ...current,
    [field]: ''
  };
  if (locatorCandidateValid(next)) {
    return next;
  };
  const order = defaultLocatorOrderForDescriptor(next).filter((kind) => !['native_ref', 'selector', 'page_binding'].includes(stringValue(kind)));
  if (order.length <= 0) {
    return null;
  }
  return locatorCandidateForOrderedKind(next, order[0]);
}

function nativeRefPrimaryLocatorCandidate(entry) {
  const current = normalizeRequestParams(entry);
  const framePath = stringValue(current.frame_path);
  const preferredKind = stringValue(current.primary_kind);
  if (preferredKind) {
    const preferredCandidate = locatorCandidateForOrderedKind(current, preferredKind);
    if (preferredCandidate) {
      return preferredCandidate;
    }
  }
  if (stringValue(current.selector)) {
    return {
      kind: 'selector',
      selector: stringValue(current.selector),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  if (stringValue(current.href)) {
    return { kind: 'href', href: stringValue(current.href), selector_index: intValue(current.selector_index), frame_path: framePath };
  }
  if (stringValue(current.role) && stringValue(current.label)) {
    return {
      kind: 'role_label',
      role: stringValue(current.role),
      tag: stringValue(current.tag),
      label: stringValue(current.label),
      type: stringValue(current.type),
      href: stringValue(current.href),
      placeholder: stringValue(current.placeholder),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  if (stringValue(current.tag) && stringValue(current.label)) {
    return {
      kind: 'tag_label',
      tag: stringValue(current.tag),
      type: stringValue(current.type),
      label: stringValue(current.label),
      href: stringValue(current.href),
      placeholder: stringValue(current.placeholder),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  if (stringValue(current.label)) {
    return {
      kind: 'label',
      tag: stringValue(current.tag),
      label: stringValue(current.label),
      type: stringValue(current.type),
      href: stringValue(current.href),
      placeholder: stringValue(current.placeholder),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  if (stringValue(current.placeholder)) {
    return {
      kind: 'placeholder',
      tag: stringValue(current.tag),
      type: stringValue(current.type),
      placeholder: stringValue(current.placeholder),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  if (stringValue(current.tag) && stringValue(current.type)) {
    return { kind: 'tag_type', tag: stringValue(current.tag), type: stringValue(current.type), selector_index: intValue(current.selector_index), frame_path: framePath };
  }
  if (stringValue(current.tag)) {
    return { kind: 'tag', tag: stringValue(current.tag), selector_index: intValue(current.selector_index), frame_path: framePath };
  }
  if (stringValue(current.type)) {
    return {
      kind: 'type',
      tag: candidateExpectedTag(current),
      type: stringValue(current.type),
      selector_index: intValue(current.selector_index),
      frame_path: framePath
    };
  }
  return null;
}

function scopeForNativeRefEntry(page, entry) {
  const framePath = stringValue(entry?.frame_path);
  if (!framePath) {
    return page;
  }
  return frameForPath(page, framePath);
}

function frameForPath(page, framePath) {
  let current = page.mainFrame();
  if (!current) {
    return null;
  }
  const normalizedPath = stringValue(framePath);
  if (!normalizedPath) {
    return current;
  }
  const rawParts = normalizedPath.split('/');
  const parts = [];
  for (const rawPart of rawParts) {
    const trimmed = stringValue(rawPart);
    if (!/^\d+$/.test(trimmed)) {
      return null;
    }
    parts.push(Number.parseInt(trimmed, 10));
  }
  for (const part of parts) {
    const children = current.childFrames();
    if (part < 0 || part >= children.length) {
      return null;
    }
    current = children[part];
    if (!current) {
      return null;
    }
  }
  return current;
}

function captureScope(params) {
  if (actionTargetRequested(params)) {
    return 'element';
  }
  if (Boolean(params.full_page)) {
    return 'full_page';
  }
  return 'viewport';
}

function waitTimeout(params) {
  const waitMs = intValue(params.wait_ms);
  if (waitMs > 0) {
    return waitMs;
  }
  return 15000;
}

function truncate(value, maxChars) {
  if (!maxChars || maxChars <= 0 || value.length <= maxChars) {
    return value;
  }
  return value.slice(0, maxChars);
}

function stringifyValue(value) {
  if (typeof value === 'string') {
    return value;
  }
  if (value === null || typeof value === 'undefined') {
    return '';
  }
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function pngSizeFromBuffer(buffer) {
  if (!Buffer.isBuffer(buffer) || buffer.length < 24) {
    return { width: 0, height: 0 };
  }
  return {
    width: buffer.readUInt32BE(16),
    height: buffer.readUInt32BE(20)
  };
}

function authorized(req) {
  if (!token) {
    return true;
  }
  const header = stringValue(req.headers.authorization);
  return header === `Bearer ${token}`;
}

function objectValue(value) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {};
  }
  return value;
}

function arrayValue(value) {
  return Array.isArray(value) ? value : [];
}

function stringSliceValue(value) {
  return arrayValue(value).map((item) => stringValue(item)).filter((item) => item !== '');
}

function normalizeRequestParams(value) {
  const input = objectValue(value);
  const out = {};
  for (const [key, current] of Object.entries(input)) {
    const normalizedKey = stringValue(key);
    if (!normalizedKey) {
      continue;
    }
    out[normalizedKey] = normalizeRequestParamValue(current);
    const snakeKey = normalizedKey
      .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
      .replace(/[-\s]+/g, '_')
      .toLowerCase();
    if (snakeKey && !(snakeKey in out)) {
      out[snakeKey] = out[normalizedKey];
    }
  }
  return out;
}

function normalizeRequestParamValue(value) {
  if (Array.isArray(value)) {
    return value.map((item) => normalizeRequestParamValue(item));
  }
  if (!value || typeof value !== 'object') {
    return value;
  }
  return normalizeRequestParams(value);
}

function buildActionResolver(page, params) {
  const elementRef = stringValue(params.element_ref);
  const directResolver = objectValue(params.element_resolver);
  if (Object.keys(directResolver).length > 0) {
    return normalizeActionResolverRequest(page, directResolver, params);
  }
  const directSelector = stringValue(params.selector);
  const defaultSource = actionResolverOrderedDescriptorSource(page, {
    element_ref: elementRef,
    selector: directSelector
  }, params, elementRef, elementRef);
  const locatorOrder = defaultLocatorOrderForDescriptor(defaultSource);
  const pageBinding = normalizePageBindingCandidate(defaultSource);
  if (locatorOrder.length === 0 && !pageBinding) {
    return null;
  }
  return normalizeActionResolverRequest(page, {
    element_ref: elementRef,
    selector: stringValue(defaultSource.selector),
    selector_index: intValue(defaultSource.selector_index),
    native_ref: stringValue(defaultSource.native_ref),
    role: stringValue(defaultSource.role),
    tag: stringValue(defaultSource.tag),
    label: stringValue(defaultSource.label),
    type: stringValue(defaultSource.type),
    href: stringValue(defaultSource.href),
    placeholder: stringValue(defaultSource.placeholder),
    page_url: stringValue(defaultSource.page_url),
    page_origin: stringValue(defaultSource.page_origin),
    page_path: stringValue(defaultSource.page_path),
    page_title: stringValue(defaultSource.page_title),
    tab_index: intValue(defaultSource.tab_index),
    locator_order: locatorOrder,
    page_binding: pageBinding
  }, params);
}

function normalizeActionResolverRequest(page, rawResolver, params) {
  const resolver = normalizeRequestParams(rawResolver);
  const resolverElementRef = stringValue(resolver.element_ref);
  const paramsElementRef = stringValue(params.element_ref);
  const resolvedSelectorIndex = firstPositiveInt(
    intValue(resolver.selector_index),
    intValue(params.selector_index),
    intValue(decodeElementRef(resolverElementRef)?.selector_index),
    intValue(decodeElementRef(paramsElementRef)?.selector_index),
    intValue(nativeRefEntryForResolver(page, resolverElementRef)?.selector_index),
    intValue(nativeRefEntryForResolver(page, paramsElementRef)?.selector_index),
    intValue(objectValue(params.element_hint).selector_index)
  );
  const resolvedFramePath = firstNonEmpty(
    stringValue(resolver.frame_path),
    stringValue(decodeElementRef(resolverElementRef)?.frame_path),
    stringValue(nativeRefEntryForResolver(page, resolverElementRef)?.frame_path),
    stringValue(params.frame_path),
    stringValue(decodeElementRef(paramsElementRef)?.frame_path),
    stringValue(nativeRefEntryForResolver(page, paramsElementRef)?.frame_path),
    stringValue(objectValue(params.element_hint).frame_path)
  );
  const resolvedSelector = firstNonEmpty(
    stringValue(resolver.selector),
    stringValue(params.selector),
    selectorFromElementRef(resolverElementRef),
    selectorFromElementRef(paramsElementRef)
  );
  const locatorOrder = normalizeLocatorOrder(arrayValue(resolver.locator_order));
  const explicitLocatorPlan = [];
  for (const candidate of arrayValue(resolver.locator_plan)) {
    appendUniqueLocatorCandidate(explicitLocatorPlan, candidate);
  }
  if (locatorOrder.length > 0) {
    const orderedPlan = locatorCandidatesFromOrderedDescriptor(
      actionResolverOrderedDescriptorSource(page, resolver, params, resolverElementRef, paramsElementRef),
      locatorOrder
    );
    if (orderedPlan.length > 0) {
      const mergedPlan = [];
      for (const candidate of orderedPlan) {
        appendUniqueLocatorCandidate(mergedPlan, candidate);
      }
      for (const candidate of explicitLocatorPlan) {
        appendUniqueLocatorCandidate(mergedPlan, candidate);
      }
      explicitLocatorPlan.length = 0;
      for (const candidate of mergedPlan) {
        explicitLocatorPlan.push(candidate);
      }
    }
  }
  const explicitMatchPlan = [];
  for (const candidate of arrayValue(resolver.match_plan)) {
    appendUniqueLocatorCandidate(explicitMatchPlan, candidate);
  }
  if (explicitMatchPlan.length === 0) {
    for (const candidate of explicitLocatorPlan) {
      if (stringValue(candidate.kind) !== 'page_binding') {
        appendUniqueLocatorCandidate(explicitMatchPlan, candidate);
      }
    }
  }
  let pageBinding = normalizePageBindingCandidate(objectValue(resolver.page_binding));
  if (!pageBinding) {
    for (const candidate of explicitLocatorPlan) {
      if (stringValue(candidate.kind) === 'page_binding') {
        pageBinding = normalizePageBindingCandidate(candidate);
        if (pageBinding) {
          break;
        }
      }
    }
  }
  if (!pageBinding) {
    pageBinding = pageBindingFromDescriptor(decodeElementRef(stringValue(resolver.element_ref))) ||
      pageBindingFromDescriptor(decodeElementRef(stringValue(params.element_ref))) ||
      pageBindingFromDescriptor(objectValue(params.element_hint));
  }
  const primaryKind = effectivePrimaryKindForResolverRequest(
    resolver,
    explicitMatchPlan,
    pageBinding
  );
  const resolutionMode = firstNonEmpty(
    stringValue(resolver.resolution_mode),
    primaryKind === 'native_ref' ? 'native_ref_first' : '',
    primaryKind === 'selector' ? 'selector_first' : 'locator_plan_only'
  );
  const elementRef = firstNonEmpty(resolverElementRef, paramsElementRef);
  const matchPlan = [];
  appendPrimaryAndSecondaryResolverCandidates(
    matchPlan,
    primaryKind,
    resolutionMode,
    elementRef,
    resolvedSelector,
    resolvedSelectorIndex,
    resolvedFramePath
  );
  for (const candidate of explicitMatchPlan) {
    appendUniqueLocatorCandidate(matchPlan, candidate);
  }
  const locatorPlan = [];
  appendPrimaryAndSecondaryResolverCandidates(
    locatorPlan,
    primaryKind,
    resolutionMode,
    elementRef,
    resolvedSelector,
    resolvedSelectorIndex,
    resolvedFramePath
  );
  for (const candidate of explicitMatchPlan) {
    appendUniqueLocatorCandidate(locatorPlan, candidate);
  }
  if (pageBinding) {
    appendUniqueLocatorCandidate(locatorPlan, pageBinding);
  }
  for (const candidate of explicitLocatorPlan) {
    appendUniqueLocatorCandidate(locatorPlan, candidate);
  }
  if (matchPlan.length === 0 && !pageBinding) {
    return null;
  }
  return {
    resolutionMode,
    primaryKind,
    elementRef,
    selector: resolvedSelector,
    framePath: resolvedFramePath,
    locatorOrder,
    locatorPlan,
    matchPlan,
    pageBinding
  };
}

function effectivePrimaryKindForResolverRequest(resolver, explicitMatchPlan, pageBinding) {
  const explicitKind = stringValue(resolver.primary_kind);
  if (explicitKind) {
    return explicitKind;
  }
  const explicitMode = stringValue(resolver.resolution_mode);
  switch (explicitMode) {
    case 'native_ref_first':
      return 'native_ref';
    case 'selector_first':
      return 'selector';
    default:
      break;
  }
  return firstNonEmpty(
    stringValue(explicitMatchPlan[0]?.kind),
    stringValue(pageBinding?.kind)
  );
}

function resolverPrimaryCandidate(primaryKind, elementRef, selector, selectorIndex, framePath) {
  switch (stringValue(primaryKind)) {
    case 'native_ref':
      if (elementRef && !decodeElementRef(elementRef)) {
        return { kind: 'native_ref', native_ref: elementRef, frame_path: stringValue(framePath) };
      }
      return null;
    case 'selector':
      if (selector) {
        return { kind: 'selector', selector, selector_index: selectorIndex, frame_path: stringValue(framePath) };
      }
      return null;
    default:
      return null;
  }
}

function appendPrimaryAndSecondaryResolverCandidates(target, primaryKind, resolutionMode, elementRef, selector, selectorIndex, framePath) {
  appendUniqueLocatorCandidate(target, resolverPrimaryCandidate(primaryKind, elementRef, selector, selectorIndex, framePath));
  const secondary = [];
  if (stringValue(primaryKind) !== 'native_ref' && elementRef && !decodeElementRef(elementRef)) {
    secondary.push({ kind: 'native_ref', native_ref: elementRef, frame_path: stringValue(framePath) });
  }
  if (stringValue(primaryKind) !== 'selector' && selector) {
    secondary.push({ kind: 'selector', selector, selector_index: selectorIndex, frame_path: stringValue(framePath) });
  }
  if (stringValue(resolutionMode) === 'selector_first' && secondary.length === 2) {
    secondary[0], secondary[1] = secondary[1], secondary[0];
  }
  for (const candidate of secondary) {
    appendUniqueLocatorCandidate(target, candidate);
  }
}

function exportActionResolver(resolver) {
  if (!resolver || typeof resolver !== 'object') {
    return null;
  }
  return {
    resolution_mode: stringValue(resolver.resolutionMode),
    primary_kind: stringValue(resolver.primaryKind),
    element_ref: stringValue(resolver.elementRef),
    selector: stringValue(resolver.selector),
    frame_path: stringValue(resolver.framePath),
    locator_order: arrayValue(resolver.locatorOrder).map((item) => stringValue(item)).filter((item) => item !== ''),
    locator_plan: arrayValue(resolver.locatorPlan)
      .map((candidate) => normalizeLocatorCandidate(candidate))
      .filter((candidate) => locatorCandidateValid(candidate)),
    match_plan: arrayValue(resolver.matchPlan)
      .map((candidate) => normalizeLocatorCandidate(candidate))
      .filter((candidate) => locatorCandidateValid(candidate)),
    page_binding: normalizePageBindingCandidate(objectValue(resolver.pageBinding))
  };
}

function actionResolverOrderedDescriptorSource(page, resolver, params, resolverElementRef, paramsElementRef) {
  const source = normalizeRequestParams(resolver);
  const resolverDescriptor = normalizeRequestParams(decodeElementRef(resolverElementRef) || {});
  const paramsDescriptor = normalizeRequestParams(decodeElementRef(paramsElementRef) || {});
  const hintDescriptor = normalizeRequestParams(objectValue(params.element_hint));
  const resolverNativeDescriptor = normalizeRequestParams(nativeRefEntryForResolver(page, resolverElementRef) || {});
  const paramsNativeDescriptor = normalizeRequestParams(nativeRefEntryForResolver(page, paramsElementRef) || {});
  if (!stringValue(source.native_ref)) {
    if (resolverElementRef && !decodeElementRef(resolverElementRef)) {
      source.native_ref = resolverElementRef;
    } else if (paramsElementRef && !decodeElementRef(paramsElementRef)) {
      source.native_ref = paramsElementRef;
    }
  }
  if (!stringValue(source.selector)) {
    source.selector = firstNonEmpty(
      stringValue(resolver.selector),
      stringValue(params.selector),
      stringValue(resolverDescriptor.selector),
      stringValue(paramsDescriptor.selector),
      stringValue(resolverNativeDescriptor.selector),
      stringValue(paramsNativeDescriptor.selector),
      selectorFromElementRef(resolverElementRef),
      selectorFromElementRef(paramsElementRef),
      stringValue(hintDescriptor.selector)
    );
  }
  if (intValue(source.selector_index) <= 0) {
    source.selector_index = firstPositiveInt(
      intValue(resolver.selector_index),
      intValue(resolverDescriptor.selector_index),
      intValue(resolverNativeDescriptor.selector_index),
      intValue(params.selector_index),
      intValue(paramsDescriptor.selector_index),
      intValue(paramsNativeDescriptor.selector_index),
      intValue(hintDescriptor.selector_index)
    );
  }
  if (!stringValue(source.frame_path)) {
    source.frame_path = firstNonEmpty(
      stringValue(resolver.frame_path),
      stringValue(resolverDescriptor.frame_path),
      stringValue(resolverNativeDescriptor.frame_path),
      stringValue(params.frame_path),
      stringValue(paramsDescriptor.frame_path),
      stringValue(paramsNativeDescriptor.frame_path),
      stringValue(hintDescriptor.frame_path)
    );
  }
  mergeOrderedDescriptorFields(source, resolverDescriptor, [
    'role',
    'tag',
    'label',
    'type',
    'href',
    'placeholder'
  ]);
  mergeOrderedDescriptorFields(source, resolverNativeDescriptor, [
    'role',
    'tag',
    'label',
    'type',
    'href',
    'placeholder'
  ]);
  mergeOrderedDescriptorFields(source, paramsDescriptor, [
    'role',
    'tag',
    'label',
    'type',
    'href',
    'placeholder'
  ]);
  mergeOrderedDescriptorFields(source, paramsNativeDescriptor, [
    'role',
    'tag',
    'label',
    'type',
    'href',
    'placeholder'
  ]);
  mergeOrderedDescriptorFields(source, hintDescriptor, [
    'role',
    'tag',
    'label',
    'type',
    'href',
    'placeholder'
  ]);
  mergeOrderedPageBindingFields(source, resolverNativeDescriptor);
  mergeOrderedPageBindingFields(source, paramsNativeDescriptor);
  const hasPageBinding = Boolean(
    stringValue(source.page_url) ||
    stringValue(source.page_origin) ||
    stringValue(source.page_path) ||
    stringValue(source.page_title) ||
    intValue(source.tab_index) > 0
  );
  if (!hasPageBinding) {
    const binding = pageBindingFromDescriptor(decodeElementRef(resolverElementRef)) ||
      pageBindingFromDescriptor(decodeElementRef(paramsElementRef)) ||
      pageBindingFromDescriptor(resolverNativeDescriptor) ||
      pageBindingFromDescriptor(paramsNativeDescriptor) ||
      pageBindingFromDescriptor(objectValue(params.element_hint));
    if (binding) {
      if (!stringValue(source.page_url)) {
        source.page_url = stringValue(binding.page_url);
      }
      if (!stringValue(source.page_origin)) {
        source.page_origin = stringValue(binding.page_origin);
      }
      if (!stringValue(source.page_path)) {
        source.page_path = stringValue(binding.page_path);
      }
      if (!stringValue(source.page_title)) {
        source.page_title = stringValue(binding.page_title);
      }
      if (intValue(source.tab_index) <= 0 && intValue(binding.tab_index) > 0) {
        source.tab_index = intValue(binding.tab_index);
      }
    }
  }
  return source;
}

function defaultLocatorOrderForDescriptor(source) {
  const current = normalizeRequestParams(source);
  const order = [];
  if (stringValue(current.native_ref)) {
    order.push('native_ref');
  }
  if (stringValue(current.selector)) {
    order.push('selector');
  }
  if (stringValue(current.href)) {
    order.push('href');
  }
  if (stringValue(current.label)) {
    if (stringValue(current.role)) {
      order.push('role_label');
    } else if (stringValue(current.tag)) {
      order.push('tag_label');
    } else {
      order.push('label');
    }
  }
  if (stringValue(current.placeholder)) {
    order.push('placeholder');
  }
  if (stringValue(current.tag) && stringValue(current.type)) {
    order.push('tag_type');
  } else if (stringValue(current.tag)) {
    order.push('tag');
  } else if (stringValue(current.type)) {
    order.push('type');
  }
  if (normalizePageBindingCandidate(current)) {
    order.push('page_binding');
  }
  return order;
}

function firstPositiveInt(...values) {
  for (const value of values) {
    const current = intValue(value);
    if (current > 0) {
      return current;
    }
  }
  return 0;
}

function nativeRefEntryForResolver(page, ref) {
  if (!page) {
    return null;
  }
  const value = stringValue(ref);
  if (!value || decodeElementRef(value)) {
    return null;
  }
  return nativeRefEntryForPage(page, value);
}

function mergeOrderedDescriptorFields(target, source, keys) {
  if (!target || !source || typeof target !== 'object' || typeof source !== 'object') {
    return;
  }
  for (const key of keys) {
    if (!stringValue(target[key]) && stringValue(source[key])) {
      target[key] = stringValue(source[key]);
    }
  }
}

function mergeOrderedPageBindingFields(target, source) {
  if (!target || !source || typeof target !== 'object' || typeof source !== 'object') {
    return;
  }
  if (!stringValue(target.page_url) && stringValue(source.page_url)) {
    target.page_url = stringValue(source.page_url);
  }
  if (!stringValue(target.page_origin) && stringValue(source.page_origin)) {
    target.page_origin = stringValue(source.page_origin);
  }
  if (!stringValue(target.page_path) && stringValue(source.page_path)) {
    target.page_path = stringValue(source.page_path);
  }
  if (!stringValue(target.page_title) && stringValue(source.page_title)) {
    target.page_title = stringValue(source.page_title);
  }
  if (intValue(target.tab_index) <= 0 && intValue(source.tab_index) > 0) {
    target.tab_index = intValue(source.tab_index);
  }
}

function normalizeLocatorOrder(values) {
  const out = [];
  const seen = new Set();
  for (const value of values) {
    const kind = stringValue(value);
    if (!kind || seen.has(kind)) {
      continue;
    }
    seen.add(kind);
    out.push(kind);
  }
  return out;
}

function locatorCandidatesFromOrderedDescriptor(source, order) {
  const normalized = normalizeRequestParams(source);
  const out = [];
  for (const kind of order) {
    const candidate = locatorCandidateForOrderedKind(normalized, kind);
    if (candidate) {
      appendUniqueLocatorCandidate(out, candidate);
    }
  }
  return out;
}

function locatorCandidateForOrderedKind(source, kind) {
  const framePath = stringValue(source.frame_path);
  switch (stringValue(kind)) {
    case 'native_ref':
      return stringValue(source.native_ref) ? { kind: 'native_ref', native_ref: stringValue(source.native_ref), frame_path: framePath } : null;
    case 'selector':
      return stringValue(source.selector)
        ? { kind: 'selector', selector: stringValue(source.selector), selector_index: intValue(source.selector_index), frame_path: framePath }
        : null;
    case 'href':
      return stringValue(source.href) ? { kind: 'href', href: stringValue(source.href), selector_index: intValue(source.selector_index), frame_path: framePath } : null;
    case 'role_label':
      return stringValue(source.role) && stringValue(source.label)
        ? {
            kind: 'role_label',
            role: stringValue(source.role),
            label: stringValue(source.label),
            tag: stringValue(source.tag),
            type: stringValue(source.type),
            href: stringValue(source.href),
            placeholder: stringValue(source.placeholder),
            selector_index: intValue(source.selector_index),
            frame_path: framePath
          }
        : null;
    case 'tag_label':
      return stringValue(source.tag) && stringValue(source.label)
        ? {
            kind: 'tag_label',
            tag: stringValue(source.tag),
            type: stringValue(source.type),
            label: stringValue(source.label),
            href: stringValue(source.href),
            placeholder: stringValue(source.placeholder),
            selector_index: intValue(source.selector_index),
            frame_path: framePath
          }
        : null;
    case 'label':
      return stringValue(source.label)
        ? {
            kind: 'label',
            label: stringValue(source.label),
            tag: stringValue(source.tag),
            type: stringValue(source.type),
            href: stringValue(source.href),
            placeholder: stringValue(source.placeholder),
            selector_index: intValue(source.selector_index),
            frame_path: framePath
          }
        : null;
    case 'placeholder':
      return stringValue(source.placeholder)
        ? {
            kind: 'placeholder',
            tag: stringValue(source.tag),
            type: stringValue(source.type),
            placeholder: stringValue(source.placeholder),
            selector_index: intValue(source.selector_index),
            frame_path: framePath
          }
        : null;
    case 'tag_type':
      return stringValue(source.tag) && stringValue(source.type)
        ? { kind: 'tag_type', tag: stringValue(source.tag), type: stringValue(source.type), selector_index: intValue(source.selector_index), frame_path: framePath }
        : null;
    case 'tag':
      return stringValue(source.tag) ? { kind: 'tag', tag: stringValue(source.tag), selector_index: intValue(source.selector_index), frame_path: framePath } : null;
    case 'type':
      return stringValue(source.type)
        ? {
            kind: 'type',
            tag: candidateExpectedTag(source),
            type: stringValue(source.type),
            selector_index: intValue(source.selector_index),
            frame_path: framePath
          }
        : null;
    case 'page_binding':
      return normalizePageBindingCandidate(source);
    default:
      return null;
  }
}

async function resolveActionLocatorWithResolver(page, resolver, params, timeout, options) {
  const outcome = {
    resolution_mode: resolver.resolutionMode,
    primary_kind: resolver.primaryKind,
    attempt_count: 0
  };
  const preferredFramePath = preferredFramePathForResolver(page, resolver);
  let blockedOutcome = null;
  let fallbackFromOutcome = null;
  if (resolver.pageBinding) {
    const blocked = await validatePageBindingForResolver(page, resolver.pageBinding, params, outcome);
    if (blocked) {
      recordResolverFailure(page, blocked);
      throw resolverHttpError(400, blocked.note || 'element ref page binding differs from current page', blocked);
    }
  }
  if (resolver.matchPlan.length === 0) {
    if (options.required) {
      throw resolverHttpError(400, 'no element matched resolver plan', normalizeResolverOutcome({
        ...outcome,
        status: 'unresolved',
        note: 'no element matched resolver plan'
      }));
    }
    return { locator: null, selector: '', outcome: normalizeResolverOutcome(outcome) };
  }
  for (let idx = 0; idx < resolver.matchPlan.length; idx += 1) {
    const candidate = resolver.matchPlan[idx];
    outcome.attempt_count += 1;
    let resolved;
    try {
      resolved = await resolveLocatorForCandidate(page, candidate, timeout, options.allowHidden, preferredFramePath);
    } catch (err) {
      throw resolverHttpError(400, errorMessage(err), normalizeResolverOutcome({
        ...outcome,
        status: 'resolution_failed',
        matched_index: idx,
        matched_kind: stringValue(candidate.kind),
        note: errorMessage(err)
      }));
    }
    if (resolved && resolved.outcome && !blockedOutcome) {
      blockedOutcome = normalizeResolverOutcome({
        ...outcome,
        matched_index: idx,
        matched_kind: stringValue(candidate.kind),
        ...resolved.outcome
      });
    }
    if (resolved && resolved.outcome) {
      fallbackFromOutcome = normalizeResolverOutcome({
        ...outcome,
        matched_index: idx,
        matched_kind: stringValue(candidate.kind),
        ...resolved.outcome
      });
    }
    if (!resolved?.locator) {
      continue;
    }
    const effectiveMatchedCandidate = effectiveMatchedCandidateForResolverMatch(candidate, resolved);
    const matchedCandidateKind = stringValue(effectiveMatchedCandidate?.kind) || stringValue(candidate.kind);
    const resolvedFramePath = stringValue(resolved?.resolved_frame_path);
    const resolvedSelectorIndex = resolvedSelectorIndexForMatchedCandidate(effectiveMatchedCandidate, resolved);
    const matchedOutcome = {
      ...outcome,
      status: 'matched',
      matched_kind: stringValue(candidate.kind),
      matched_index: idx,
      matched_candidate_kind: matchedCandidateKind,
      note: resolverMatchedNote(effectiveMatchedCandidate)
    };
    if (resolvedFramePath) {
      matchedOutcome.resolved_frame_path = resolvedFramePath;
    }
    if (resolvedSelectorIndex > 0) {
      matchedOutcome.resolved_selector_index = resolvedSelectorIndex;
    }
    if (fallbackFromOutcome) {
      matchedOutcome.fallback_from_kind = stringValue(fallbackFromOutcome.matched_kind) || stringValue(fallbackFromOutcome.candidate_kind);
      matchedOutcome.fallback_from_index = intValue(fallbackFromOutcome.matched_index);
      matchedOutcome.fallback_from_blocked_by = stringValue(fallbackFromOutcome.blocked_by);
      matchedOutcome.fallback_from_ambiguity_class = stringValue(fallbackFromOutcome.ambiguity_class);
      matchedOutcome.fallback_from_candidate_strength = stringValue(fallbackFromOutcome.candidate_strength);
      matchedOutcome.fallback_from_manual_retry_hint = stringValue(fallbackFromOutcome.manual_retry_hint);
      matchedOutcome.fallback_from_specificity_fields = stringSliceValue(fallbackFromOutcome.specificity_fields);
    }
    if (maybeRebindResolverNativeRef(page, resolver, effectiveMatchedCandidate, resolved)) {
      matchedOutcome.native_ref_rebound = true;
    }
    return {
      locator: resolved.locator,
      selector: stringValue(candidate.selector),
      outcome: normalizeResolverOutcome(matchedOutcome)
    };
  }
  const finalOutcome = preferredFailureOutcomeForResolver(blockedOutcome, fallbackFromOutcome);
  if (finalOutcome) {
    recordResolverFailure(page, finalOutcome);
    throw resolverHttpError(400, finalOutcome.note || 'element ref no longer matches current frame tree', finalOutcome);
  }
  throw resolverHttpError(400, 'no element matched resolver plan', normalizeResolverOutcome({
    ...outcome,
    status: 'unresolved',
    note: 'no element matched resolver plan'
  }));
}

function preferredFailureOutcomeForResolver(blockedOutcome, fallbackFromOutcome) {
  if (!blockedOutcome) {
    return fallbackFromOutcome || null;
  }
  if (!fallbackFromOutcome || fallbackFromOutcome === blockedOutcome) {
    return blockedOutcome;
  }
  if (
    stringValue(blockedOutcome.status) === 'page_binding_blocked' &&
    stringValue(blockedOutcome.blocked_by) === 'frame_path' &&
    !(
      stringValue(fallbackFromOutcome.status) === 'page_binding_blocked' &&
      stringValue(fallbackFromOutcome.blocked_by) === 'frame_path'
    )
  ) {
    return fallbackFromOutcome;
  }
  return blockedOutcome;
}

function preferredFramePathForResolver(page, resolver) {
  if (resolver && typeof resolver === 'object') {
    const directFramePath = stringValue(resolver.framePath || resolver.frame_path);
    if (directFramePath) {
      return directFramePath;
    }
  }
  const refs = [];
  if (resolver && typeof resolver === 'object') {
    const directRef = stringValue(resolver.elementRef || resolver.element_ref);
    if (directRef) {
      refs.push(directRef);
    }
    for (const candidate of arrayValue(resolver.matchPlan || resolver.match_plan)) {
      if (stringValue(candidate?.kind) === 'native_ref' && stringValue(candidate?.native_ref)) {
        refs.push(stringValue(candidate.native_ref));
      }
    }
    for (const candidate of arrayValue(resolver.locatorPlan || resolver.locator_plan)) {
      if (stringValue(candidate?.kind) === 'native_ref' && stringValue(candidate?.native_ref)) {
        refs.push(stringValue(candidate.native_ref));
      }
    }
    for (const source of [resolver.matchPlan || resolver.match_plan, resolver.locatorPlan || resolver.locator_plan]) {
      for (const candidate of arrayValue(source)) {
        const framePath = stringValue(candidate?.frame_path || candidate?.framePath);
        if (framePath) {
          return framePath;
        }
      }
    }
  }
  for (const ref of refs) {
    const entry = nativeRefEntryForResolver(page, ref);
    if (entry && stringValue(entry.frame_path)) {
      return stringValue(entry.frame_path);
    }
  }
  return '';
}

function nativeRefForResolverRebinding(resolver) {
  if (!resolver || typeof resolver !== 'object') {
    return '';
  }
  const directRef = stringValue(resolver.elementRef || resolver.element_ref);
  if (directRef) {
    return directRef;
  }
  for (const source of [resolver.matchPlan || resolver.match_plan, resolver.locatorPlan || resolver.locator_plan]) {
    for (const candidate of arrayValue(source)) {
      if (stringValue(candidate?.kind) === 'native_ref' && stringValue(candidate?.native_ref)) {
        return stringValue(candidate.native_ref);
      }
    }
  }
  return '';
}

function maybeRebindResolverNativeRef(page, resolver, matchedCandidate, resolved) {
  if (!page || !resolver || typeof resolver !== 'object') {
    return false;
  }
  if (stringValue(resolver.primaryKind || resolver.primary_kind) !== 'native_ref') {
    return false;
  }
  const effectiveMatchedCandidate = normalizeLocatorCandidate(objectValue(matchedCandidate));
  const matchedKind = stringValue(effectiveMatchedCandidate?.kind);
  if (!matchedKind || matchedKind === 'native_ref' || matchedKind === 'page_binding') {
    return false;
  }
  const nativeRef = nativeRefForResolverRebinding(resolver);
  if (!nativeRef) {
    return false;
  }
  const refs = nativeRefStoreForPage(page);
  const current = refs.get(nativeRef);
  if (!current) {
    return false;
  }
  refs.set(nativeRef, reboundNativeRefEntryForMatchedCandidate(current, effectiveMatchedCandidate, resolved));
  return true;
}

function effectiveMatchedCandidateForResolverMatch(matchedCandidate, resolved) {
  const internalMatchedCandidate = normalizeLocatorCandidate(objectValue(resolved?.matched_candidate));
  if (locatorCandidateValid(internalMatchedCandidate)) {
    return internalMatchedCandidate;
  }
  return normalizeLocatorCandidate(objectValue(matchedCandidate));
}

function reboundNativeRefEntryForMatchedCandidate(entry, candidate, resolved) {
  const next = { ...entry };
  if (resolved && typeof resolved === 'object' && Object.prototype.hasOwnProperty.call(resolved, 'resolved_frame_path')) {
    next.frame_path = stringValue(resolved.resolved_frame_path);
  }
  if (stringValue(candidate?.kind)) {
    next.primary_kind = stringValue(candidate.kind);
  }
  const resolvedSelectorIndex = resolvedSelectorIndexForMatchedCandidate(candidate, resolved);
  switch (stringValue(candidate?.kind)) {
    case 'selector':
      next.selector = stringValue(candidate?.selector);
      next.selector_index = resolvedSelectorIndex;
      break;
    case 'href':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = stringValue(candidate?.href) || stringValue(next.href);
      break;
    case 'role_label':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = stringValue(candidate?.href);
      next.role = stringValue(candidate?.role) || stringValue(next.role);
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      next.label = stringValue(candidate?.label) || stringValue(next.label);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      next.placeholder = stringValue(candidate?.placeholder) || stringValue(next.placeholder);
      break;
    case 'tag_label':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = stringValue(candidate?.href);
      next.role = '';
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      next.label = stringValue(candidate?.label) || stringValue(next.label);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      next.placeholder = stringValue(candidate?.placeholder) || stringValue(next.placeholder);
      break;
    case 'label':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = stringValue(candidate?.href);
      next.role = '';
      next.label = stringValue(candidate?.label) || stringValue(next.label);
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      next.placeholder = stringValue(candidate?.placeholder) || stringValue(next.placeholder);
      break;
    case 'placeholder':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = '';
      next.role = '';
      next.label = '';
      next.placeholder = stringValue(candidate?.placeholder) || stringValue(next.placeholder);
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      break;
    case 'tag_type':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = '';
      next.role = '';
      next.label = '';
      next.placeholder = '';
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      break;
    case 'tag':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = '';
      next.role = '';
      next.label = '';
      next.placeholder = '';
      next.type = '';
      next.tag = stringValue(candidate?.tag) || stringValue(next.tag);
      break;
    case 'type':
      next.selector = '';
      next.selector_index = resolvedSelectorIndex;
      next.href = '';
      next.role = '';
      next.label = '';
      next.placeholder = '';
      next.tag = stringValue(candidate?.tag) || candidateExpectedTag(candidate) || stringValue(next.tag);
      next.type = stringValue(candidate?.type) || stringValue(next.type);
      break;
    default:
      break;
  }
  return next;
}

function resolvedSelectorIndexForMatchedCandidate(candidate, resolved) {
  if (resolved && typeof resolved === 'object' && Object.prototype.hasOwnProperty.call(resolved, 'resolved_selector_index')) {
    return intValue(resolved.resolved_selector_index);
  }
  return intValue(candidate?.selector_index);
}

function recordResolverFailure(page, outcome) {
  if (!page || !outcome || typeof outcome !== 'object') {
    return;
  }
  const status = stringValue(outcome.status);
  const blockedBy = stringValue(outcome.blocked_by);
  if (status === 'unresolved' && (
    blockedBy === 'multiple_candidates' ||
    blockedBy === 'multiple_candidates_same_semantic' ||
    blockedBy === 'multiple_candidates_filtered' ||
    blockedBy === 'selector_index' ||
    blockedBy === 'selector_index_out_of_range' ||
    blockedBy === 'selector_index_filtered_out'
  )) {
    const recoveryAction = stringValue(outcome.recovery_action) || 'browser action=snapshot';
    pushPageEvent(
      page,
      'resolver',
      'ambiguous_target',
      stringValue(outcome.note) || 'resolver fallback is ambiguous after DOM reorder',
      safeCall(() => page.url()),
      {
        resolver_status: status,
        candidate_kind: stringValue(outcome.candidate_kind),
        candidate_strength: stringValue(outcome.candidate_strength),
        ambiguity_class: stringValue(outcome.ambiguity_class),
        retry_disposition: stringValue(outcome.retry_disposition),
        manual_retry_hint: stringValue(outcome.manual_retry_hint) || manualRetryHintForBlockedBy(blockedBy, recoveryAction),
        next_step_alias: stringValue(outcome.next_step_alias) || nextStepAliasForRecoveryAction(recoveryAction),
        blocked_by: blockedBy,
        locator_count: intValue(outcome.locator_count),
        candidate_count: intValue(outcome.candidate_count),
        preferred_ordinal: intValue(outcome.preferred_ordinal),
        specificity_fields: stringSliceValue(outcome.specificity_fields),
        recovery_action: recoveryAction
      }
    );
    return;
  }
  const candidateCount = intValue(outcome.candidate_count);
  const staleTargetFailure = status === 'page_binding_blocked' ||
    (status === 'unresolved' && blockedBy === 'frame_path' && candidateCount <= 0);
  if (!staleTargetFailure) {
    return;
  }
  const message = stringValue(outcome.note) || (
    blockedBy === 'frame_path'
      ? 'resolver target no longer matches the preferred frame path'
      : 'element ref belongs to a different page'
  );
  const recoveryAction = stringValue(outcome.recovery_action) || 'browser action=snapshot';
  const state = pageSessionStateForPage(page);
  state.staleTarget = {
    status,
    blockedBy,
    message,
    recoveryAction,
    occurredAt: Date.now()
  };
  state.updatedAt = Date.now();
  pushPageEvent(
    page,
    'resolver',
    'stale_target',
    message,
    safeCall(() => page.url()),
    {
      resolver_status: status,
      blocked_by: blockedBy,
      recovery_action: recoveryAction
    }
  );
}

function clearStaleTargetState(page) {
  if (!page) {
    return;
  }
  const state = pageSessionStateForPage(page);
  if (!state || !state.staleTarget) {
    return;
  }
  state.staleTarget = null;
  state.updatedAt = Date.now();
}

async function validatePageBindingForResolver(page, pageBinding, params, baseOutcome) {
  const currentURL = stringValue(page.url());
  const currentTitle = stringValue(await page.title());
  const requestedTabIndex = intValue(params.tab_index);
  const expectedURL = stringValue(pageBinding.page_url);
  const expectedOrigin = stringValue(pageBinding.page_origin);
  const expectedPath = stringValue(pageBinding.page_path);
  const expectedTitle = stringValue(pageBinding.page_title);
  const expectedTabIndex = intValue(pageBinding.tab_index);
  let blockedBy = pageBindingBlockedByForResolver(
    expectedURL,
    expectedOrigin,
    expectedPath,
    expectedTitle,
    expectedTabIndex,
    currentURL,
    currentTitle,
    requestedTabIndex
  );
  const expectedFramePath = stringValue(pageBinding.frame_path);
  if (!blockedBy && expectedFramePath && !pageHasDescendantFramePath(page, expectedFramePath)) {
    blockedBy = 'frame_path';
  }
  if (!blockedBy) {
    return null;
  }
  if (blockedBy === 'frame_path') {
    return normalizeResolverOutcome({
      ...baseOutcome,
      status: 'page_binding_blocked',
      blocked_by: blockedBy,
      note: `element ref expects frame path ${expectedFramePath} but current frame tree no longer matches`
    });
  }
  const expected = firstNonEmpty(
    comparableURL(expectedURL),
    [expectedOrigin, expectedPath, expectedTitle].filter(Boolean).join(' ')
  );
  const current = firstNonEmpty(
    comparableURL(currentURL),
    [urlOrigin(currentURL), urlPath(currentURL), currentTitle].filter(Boolean).join(' ')
  );
  return normalizeResolverOutcome({
    ...baseOutcome,
    status: 'page_binding_blocked',
    blocked_by: blockedBy,
    note: `element ref expects ${expected || 'the original page'} but current page is ${current || 'unknown'}`
  });
}

function pageHasDescendantFramePath(page, expectedFramePath) {
  const expected = stringValue(expectedFramePath);
  if (!expected) {
    return true;
  }
  return descendantFrameEntriesForPage(page).some((entry) => stringValue(entry.path) === expected);
}

function pageBindingBlockedByForResolver(expectedURL, expectedOrigin, expectedPath, expectedTitle, expectedTabIndex, currentURL, currentTitle, requestedTabIndex) {
  const hasBinding = Boolean(expectedURL || expectedOrigin || expectedPath || expectedTitle || expectedTabIndex > 0);
  if (!hasBinding) {
    return '';
  }
  if (expectedURL) {
    const normalizedExpected = comparableURL(expectedURL);
    const normalizedCurrent = comparableURL(currentURL);
    if (normalizedExpected && normalizedCurrent) {
      if (normalizedExpected !== normalizedCurrent) {
        return 'page_url';
      }
    } else if (stringValue(expectedURL) !== stringValue(currentURL)) {
      return 'page_url';
    }
  }
  if (!expectedURL) {
    const currentOrigin = urlOrigin(currentURL);
    if (expectedOrigin && expectedOrigin !== currentOrigin) {
      return 'page_origin';
    }
    const currentPath = urlPath(currentURL);
    if (expectedPath && expectedPath !== currentPath) {
      return 'page_path';
    }
    if (expectedTitle && expectedTitle !== currentTitle) {
      return 'page_title';
    }
  }
  if (expectedTabIndex > 0 && requestedTabIndex > 0 && expectedTabIndex !== requestedTabIndex) {
    return 'tab_index';
  }
  return '';
}

async function resolveLocatorForCandidate(page, candidate, timeout, allowHidden, preferredFramePath) {
  const preferredPath = firstNonEmpty(
    stringValue(candidate?.frame_path || candidate?.framePath),
    stringValue(preferredFramePath)
  );
  const preferredIndex = preferredOrdinalForLocatorCandidate(candidate);
  if (!preferredPath && preferredIndex <= 0 && candidateRequiresUniqueMatchWithoutOrdinal(candidate) && candidateAllowsFrameTraversal(candidate)) {
    return await resolveUniqueTraversalLocatorForCandidate(page, candidate, timeout, allowHidden);
  }
  const resolved = locatorForCandidate(page, candidate);
  const effectiveCandidate = effectiveLocatorCandidateForResolvedCandidate(candidate, resolved);
  const direct = await firstUsableLocator(resolved?.locator || null, timeout, allowHidden, effectiveCandidate);
  if (direct) {
    return {
      locator: direct.locator,
      outcome: resolved?.outcome || null,
      resolved_frame_path: stringValue(resolved?.frame_path),
      resolved_selector_index: intValue(direct.resolved_selector_index),
      matched_candidate: effectiveCandidate
    };
  }
  let blockedOutcome = (await candidateWeakFallbackBlockedOutcome(resolved?.locator || null, effectiveCandidate)) || resolved?.outcome || null;
  for (const fallbackCandidate of arrayValue(resolved?.fallback_candidates)) {
    const fallbackLocator = locatorForCandidateInScope(resolved?.scope || page, fallbackCandidate);
    const fallbackMatched = await firstUsableLocator(fallbackLocator, timeout, allowHidden, fallbackCandidate);
    if (fallbackMatched) {
      return {
        locator: fallbackMatched.locator,
        outcome: blockedOutcome,
        resolved_frame_path: stringValue(resolved?.frame_path),
        resolved_selector_index: intValue(fallbackMatched.resolved_selector_index),
        matched_candidate: fallbackCandidate
      };
    }
    if (!blockedOutcome) {
      blockedOutcome = await candidateWeakFallbackBlockedOutcome(fallbackLocator, fallbackCandidate);
    }
  }
  if (!candidateAllowsFrameTraversal(candidate)) {
    return {
      locator: null,
      outcome: blockedOutcome
    };
  }
  if (preferredPath && preferredIndex <= 0 && candidateRequiresUniqueMatchWithoutOrdinal(candidate)) {
    return await resolvePreferredUniqueTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredPath, blockedOutcome);
  }
  if (preferredPath && preferredIndex > 0) {
    if (candidateAllowsNearestOrdinalFallback(candidate)) {
      return await resolvePreferredNearestTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredPath, blockedOutcome);
    }
    return await resolvePreferredExactOrdinalTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredPath, blockedOutcome);
  }
  for (const frame of descendantFramesForPage(page, preferredFramePath)) {
    const scoped = locatorForCandidateInScope(frame, candidate);
    const matched = await firstUsableLocator(scoped, timeout, allowHidden, candidate);
    if (matched) {
      return {
        locator: matched.locator,
        outcome: blockedOutcome,
        resolved_frame_path: framePathForFrame(frame, safeCall(() => page.mainFrame())),
        resolved_selector_index: intValue(matched.resolved_selector_index)
      };
    }
    if (!blockedOutcome) {
      blockedOutcome = await candidateWeakFallbackBlockedOutcome(scoped, candidate);
    }
  }
  return {
    locator: null,
    outcome: blockedOutcome
  };
}

function effectiveLocatorCandidateForResolvedCandidate(candidate, resolved) {
  if (resolved && typeof resolved === 'object') {
    const resolvedCandidate = normalizeLocatorCandidate(objectValue(resolved.candidate));
    if (locatorCandidateValid(resolvedCandidate)) {
      return resolvedCandidate;
    }
  }
  return candidate;
}

async function resolvePreferredExactOrdinalTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredFramePath, blockedOutcome) {
  const preferredIndex = preferredOrdinalForLocatorCandidate(candidate);
  const groups = rankedDescendantFrameEntryGroupsForPage(page, preferredFramePath);
  let totalLocatorCount = 0;
  for (const group of groups) {
    let groupLocatorCount = 0;
    let groupCandidateCount = 0;
    let matchedLocator = null;
    let matchedFramePath = '';
    let matchedSelectorIndex = -1;
    for (const entry of group) {
      const scoped = locatorForCandidateInScope(entry.frame, candidate);
      const count = await scoped.count().catch(() => 0);
      if (count <= 0) {
        continue;
      }
      groupLocatorCount += count;
      totalLocatorCount += count;
      if (!blockedOutcome) {
        blockedOutcome = await candidateWeakFallbackBlockedOutcome(scoped, candidate);
      }
      const matchingIndices = await locatorCandidateMatchingIndices(scoped, candidate);
      if (!matchingIndices.includes(preferredIndex)) {
        continue;
      }
      const current = scoped.nth(preferredIndex);
      try {
        await current.waitFor({ state: allowHidden ? 'attached' : 'visible', timeout });
      } catch {
        continue;
      }
      groupCandidateCount += 1;
      if (!matchedLocator) {
        matchedLocator = current;
        matchedFramePath = stringValue(entry.path);
        matchedSelectorIndex = preferredIndex;
      } else {
        matchedLocator = null;
        matchedFramePath = '';
        matchedSelectorIndex = -1;
      }
    }
    if (groupCandidateCount === 1 && matchedLocator) {
      return {
        locator: matchedLocator,
        outcome: blockedOutcome,
        resolved_frame_path: matchedFramePath,
        resolved_selector_index: matchedSelectorIndex
      };
    }
    if (groupCandidateCount > 1) {
      const kind = stringValue(candidate?.kind) || 'semantic';
      const blockedBy = candidateSpecificityFields(candidate).length > 0 ? 'multiple_candidates_filtered' : 'multiple_candidates_same_semantic';
      return {
        locator: null,
        outcome: {
          status: 'unresolved',
          candidate_kind: kind,
          candidate_strength: candidateStrength(candidate),
          ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
          retry_disposition: retryDispositionForBlockedBy(blockedBy),
          manual_retry_hint: 'add_specificity',
          next_step_alias: 'snapshot',
          blocked_by: blockedBy,
          locator_count: groupLocatorCount,
          candidate_count: groupCandidateCount,
          preferred_ordinal: preferredIndex,
          specificity_fields: candidateSpecificityFields(candidate),
          note: blockedBy === 'multiple_candidates_filtered'
            ? `${kind} fallback narrowed ${groupLocatorCount} locator candidate(s) within preferred frame group to ${groupCandidateCount} exact-ordinal candidate(s), but a unique match is still required`
            : `${kind} fallback found ${groupCandidateCount} exact-ordinal candidate(s) within preferred frame group, but a unique match is still required`
        }
      };
    }
  }
  if (!blockedOutcome) {
    blockedOutcome = preferredFrameGroupMissOutcome(candidate, totalLocatorCount);
  }
  return {
    locator: null,
    outcome: blockedOutcome
  };
}

async function resolvePreferredNearestTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredFramePath, blockedOutcome) {
  const preferredIndex = preferredOrdinalForLocatorCandidate(candidate);
  const groups = rankedDescendantFrameEntryGroupsForPage(page, preferredFramePath);
  let totalLocatorCount = 0;
  for (const group of groups) {
    let groupLocatorCount = 0;
    let bestDistance = Number.POSITIVE_INFINITY;
    let bestCandidateCount = 0;
    let matchedLocator = null;
    let matchedFramePath = '';
    let matchedSelectorIndex = -1;
    for (const entry of group) {
      const scoped = locatorForCandidateInScope(entry.frame, candidate);
      const count = await scoped.count().catch(() => 0);
      if (count <= 0) {
        continue;
      }
      groupLocatorCount += count;
      totalLocatorCount += count;
      if (!blockedOutcome) {
        blockedOutcome = await candidateWeakFallbackBlockedOutcome(scoped, candidate);
      }
      const matchingIndices = await locatorCandidateMatchingIndices(scoped, candidate);
      if (matchingIndices.length <= 0) {
        continue;
      }
      const bestIndex = nearestMatchingIndexForPreferredOrdinal(matchingIndices, preferredIndex);
      if (bestIndex < 0) {
        continue;
      }
      const current = scoped.nth(bestIndex);
      try {
        await current.waitFor({ state: allowHidden ? 'attached' : 'visible', timeout });
      } catch {
        continue;
      }
      const distance = Math.abs(bestIndex - preferredIndex);
      if (distance < bestDistance) {
        bestDistance = distance;
        bestCandidateCount = 1;
        matchedLocator = current;
        matchedFramePath = stringValue(entry.path);
        matchedSelectorIndex = bestIndex;
        continue;
      }
      if (distance === bestDistance) {
        bestCandidateCount += 1;
        matchedLocator = null;
        matchedFramePath = '';
        matchedSelectorIndex = -1;
      }
    }
    if (bestCandidateCount === 1 && matchedLocator) {
      return {
        locator: matchedLocator,
        outcome: blockedOutcome,
        resolved_frame_path: matchedFramePath,
        resolved_selector_index: matchedSelectorIndex
      };
    }
    if (bestCandidateCount > 1) {
      const kind = stringValue(candidate?.kind) || 'semantic';
      const blockedBy = candidateSpecificityFields(candidate).length > 0 ? 'multiple_candidates_filtered' : 'multiple_candidates_same_semantic';
      return {
        locator: null,
        outcome: {
          status: 'unresolved',
          candidate_kind: kind,
          candidate_strength: candidateStrength(candidate),
          ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
          retry_disposition: retryDispositionForBlockedBy(blockedBy),
          manual_retry_hint: 'add_specificity',
          next_step_alias: 'snapshot',
          blocked_by: blockedBy,
          locator_count: groupLocatorCount,
          candidate_count: bestCandidateCount,
          preferred_ordinal: preferredIndex,
          specificity_fields: candidateSpecificityFields(candidate),
          note: blockedBy === 'multiple_candidates_filtered'
            ? `${kind} fallback narrowed ${groupLocatorCount} locator candidate(s) within preferred frame group to ${bestCandidateCount} equally-ranked preferred-ordinal candidate(s), but a unique match is still required`
            : `${kind} fallback found ${bestCandidateCount} equally-ranked preferred-ordinal candidate(s) within preferred frame group, but a unique match is still required`
        }
      };
    }
    if (!blockedOutcome && groupLocatorCount > 0) {
      blockedOutcome = preferredFrameGroupMissOutcome(candidate, groupLocatorCount);
    }
  }
  if (!blockedOutcome) {
    blockedOutcome = preferredFrameGroupMissOutcome(candidate, totalLocatorCount);
  }
  return {
    locator: null,
    outcome: blockedOutcome
  };
}

function preferredFrameGroupMissOutcome(candidate, locatorCount) {
  const kind = stringValue(candidate?.kind) || 'semantic';
  return {
    status: 'unresolved',
    candidate_kind: kind,
    candidate_strength: candidateStrength(candidate),
    retry_disposition: retryDispositionForBlockedBy('frame_path'),
    manual_retry_hint: manualRetryHintForBlockedBy('frame_path', 'browser action=snapshot'),
    next_step_alias: 'snapshot',
    blocked_by: 'frame_path',
    locator_count: locatorCount,
    candidate_count: 0,
    specificity_fields: candidateSpecificityFields(candidate),
    note: locatorCount > 0
      ? `${kind} fallback found no matching candidate within the preferred frame path`
      : `${kind} fallback could not find any candidate inside the preferred frame path`
  };
}

async function resolvePreferredUniqueTraversalLocatorForCandidate(page, candidate, timeout, allowHidden, preferredFramePath, blockedOutcome) {
  const groups = rankedDescendantFrameEntryGroupsForPage(page, preferredFramePath);
  let totalLocatorCount = 0;
  for (const group of groups) {
    let groupLocatorCount = 0;
    let groupCandidateCount = 0;
    let matchedLocator = null;
    let matchedFramePath = '';
    let matchedSelectorIndex = -1;
    for (const entry of group) {
      const scoped = locatorForCandidateInScope(entry.frame, candidate);
      const count = await scoped.count().catch(() => 0);
      if (count <= 0) {
        continue;
      }
      groupLocatorCount += count;
      totalLocatorCount += count;
      if (!blockedOutcome) {
        blockedOutcome = await candidateWeakFallbackBlockedOutcome(scoped, candidate);
      }
      const matchingIndices = await locatorCandidateMatchingIndices(scoped, candidate);
      groupCandidateCount += matchingIndices.length;
      if (matchingIndices.length !== 1) {
        continue;
      }
      const current = scoped.nth(matchingIndices[0]);
      try {
        await current.waitFor({ state: allowHidden ? 'attached' : 'visible', timeout });
        if (!matchedLocator) {
          matchedLocator = current;
          matchedFramePath = stringValue(entry.path);
          matchedSelectorIndex = matchingIndices[0];
        }
      } catch {
        continue;
      }
    }
    if (groupCandidateCount === 1 && matchedLocator) {
      return {
        locator: matchedLocator,
        outcome: blockedOutcome,
        resolved_frame_path: matchedFramePath,
        resolved_selector_index: matchedSelectorIndex
      };
    }
    if (groupCandidateCount > 1) {
      const kind = stringValue(candidate?.kind) || 'semantic';
      const blockedBy = candidateSpecificityFields(candidate).length > 0 ? 'multiple_candidates_filtered' : 'multiple_candidates_same_semantic';
      return {
        locator: null,
        outcome: {
          status: 'unresolved',
          candidate_kind: kind,
          candidate_strength: candidateStrength(candidate),
          ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
          retry_disposition: retryDispositionForBlockedBy(blockedBy),
          manual_retry_hint: manualRetryHintForBlockedBy(blockedBy, 'browser action=snapshot'),
          next_step_alias: 'snapshot',
          blocked_by: blockedBy,
          locator_count: groupLocatorCount,
          candidate_count: groupCandidateCount,
          specificity_fields: candidateSpecificityFields(candidate),
          note: blockedBy === 'multiple_candidates_filtered'
            ? `${kind} fallback narrowed ${groupLocatorCount} locator candidate(s) within preferred frame group to ${groupCandidateCount} matching candidate(s), but a unique match is still required`
            : `${kind} fallback requires a unique same-semantic match within preferred frame group but ${groupCandidateCount} matching candidate(s) remain`
        }
      };
    }
    if (!blockedOutcome && groupLocatorCount > 0) {
      blockedOutcome = preferredFrameGroupMissOutcome(candidate, groupLocatorCount);
    }
  }
  if (!blockedOutcome) {
    blockedOutcome = preferredFrameGroupMissOutcome(candidate, totalLocatorCount);
  }
  return {
    locator: null,
    outcome: blockedOutcome
  };
}

async function resolveUniqueTraversalLocatorForCandidate(page, candidate, timeout, allowHidden) {
  const scopes = [];
  const resolved = locatorForCandidate(page, candidate);
  scopes.push({
    locator: resolved?.locator || null,
    framePath: stringValue(resolved?.frame_path)
  });
  let blockedOutcome = (await candidateWeakFallbackBlockedOutcome(resolved?.locator || null, candidate)) || resolved?.outcome || null;
  for (const entry of descendantFrameEntriesForPage(page)) {
    scopes.push({
      locator: locatorForCandidateInScope(entry.frame, candidate),
      framePath: stringValue(entry.path)
    });
  }
  let totalLocatorCount = 0;
  let totalCandidateCount = 0;
  let matchedLocator = null;
  let matchedSelectorIndex = -1;
  let matchedFramePath = '';
  for (const scope of scopes) {
    const locator = scope?.locator || null;
    if (!locator) {
      continue;
    }
    const count = await locator.count().catch(() => 0);
    if (count <= 0) {
      continue;
    }
    totalLocatorCount += count;
    if (!blockedOutcome) {
      blockedOutcome = await candidateWeakFallbackBlockedOutcome(locator, candidate);
    }
    const matchingIndices = await locatorCandidateMatchingIndices(locator, candidate);
    totalCandidateCount += matchingIndices.length;
    if (matchingIndices.length !== 1) {
      continue;
    }
    const current = locator.nth(matchingIndices[0]);
    try {
      await current.waitFor({ state: allowHidden ? 'attached' : 'visible', timeout });
      if (!matchedLocator) {
        matchedLocator = current;
        matchedSelectorIndex = matchingIndices[0];
        matchedFramePath = stringValue(scope.framePath);
      }
    } catch {
      continue;
    }
  }
  if (totalCandidateCount === 1 && matchedLocator) {
    return {
      locator: matchedLocator,
      outcome: blockedOutcome,
      resolved_frame_path: matchedFramePath,
      resolved_selector_index: matchedSelectorIndex
    };
  }
  if (totalCandidateCount > 1) {
    const kind = stringValue(candidate?.kind) || 'semantic';
    const blockedBy = candidateSpecificityFields(candidate).length > 0 ? 'multiple_candidates_filtered' : 'multiple_candidates_same_semantic';
    return {
      locator: null,
      outcome: {
        status: 'unresolved',
        candidate_kind: kind,
        candidate_strength: candidateStrength(candidate),
        ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
        retry_disposition: retryDispositionForBlockedBy(blockedBy),
        manual_retry_hint: manualRetryHintForBlockedBy(blockedBy, 'browser action=snapshot'),
        next_step_alias: 'snapshot',
        blocked_by: blockedBy,
        locator_count: totalLocatorCount,
        candidate_count: totalCandidateCount,
        specificity_fields: candidateSpecificityFields(candidate),
        note: blockedBy === 'multiple_candidates_filtered'
          ? `${kind} fallback narrowed ${totalLocatorCount} locator candidate(s) across frame tree to ${totalCandidateCount} matching candidate(s), but a unique match is still required`
          : `${kind} fallback requires a unique same-semantic match across frame tree but ${totalCandidateCount} matching candidate(s) remain`
      }
    };
  }
  return {
    locator: null,
    outcome: blockedOutcome
  };
}

async function firstUsableLocator(locator, timeout, allowHidden, candidate) {
  if (!locator) {
    return null;
  }
  const count = await locator.count().catch(() => 0);
  if (count <= 0) {
    return null;
  }
  const preferredIndex = preferredOrdinalForLocatorCandidate(candidate);
  if (preferredIndex <= 0 && candidateRequiresUniqueMatchWithoutOrdinal(candidate)) {
    const matchingIndices = await locatorCandidateMatchingIndices(locator, candidate);
    if (matchingIndices.length !== 1) {
      return null;
    }
    const current = locator.nth(matchingIndices[0]);
    return usableActionLocator(current, timeout, allowHidden, matchingIndices[0]);
  }
  let attachedFallback = null;
  for (const idx of locatorCandidateScanOrder(count, preferredIndex, candidate)) {
    const current = locator.nth(idx);
    if (!(await locatorMatchesCandidate(current, candidate))) {
      continue;
    }
    const usable = await usableActionLocator(current, timeout, allowHidden, idx);
    if (!usable) {
      continue;
    }
    if (!usable.attached_visibility_fallback) {
      return usable;
    }
    if (!attachedFallback) {
      attachedFallback = usable;
    }
  }
  return attachedFallback;
}

async function usableActionLocator(locator, timeout, allowHidden, resolvedSelectorIndex) {
  try {
    await locator.waitFor({ state: allowHidden ? 'attached' : 'visible', timeout });
    return {
      locator,
      resolved_selector_index: resolvedSelectorIndex
    };
  } catch (err) {
    if (allowHidden || !timeoutErrorLike(err)) {
      return null;
    }
  }
  try {
    await locator.waitFor({ state: 'attached', timeout: Math.max(1, Math.min(timeout, 500)) });
    return {
      locator,
      resolved_selector_index: resolvedSelectorIndex,
      attached_visibility_fallback: true
    };
  } catch {
    return null;
  }
}

async function locatorCandidateMatchingIndices(locator, candidate) {
  if (!locator) {
    return [];
  }
  const count = await locator.count().catch(() => 0);
  if (count <= 0) {
    return [];
  }
  const matches = [];
  for (let idx = 0; idx < count; idx += 1) {
    const current = locator.nth(idx);
    if (await locatorMatchesCandidate(current, candidate)) {
      matches.push(idx);
    }
  }
  return matches;
}

function preferredOrdinalForLocatorCandidate(candidate) {
  const kind = stringValue(candidate?.kind);
  if (!kind || kind === 'selector' || kind === 'native_ref' || kind === 'page_binding') {
    return -1;
  }
  return intValue(candidate?.selector_index);
}

function candidateAllowsNearestOrdinalFallback(candidate) {
  switch (stringValue(candidate?.kind)) {
    case 'href':
    case 'role_label':
    case 'tag_label':
    case 'label':
      return true;
    case 'placeholder':
    case 'tag_type':
    case 'tag':
    case 'type':
      return false;
    default:
      return true;
  }
}

function candidateRequiresUniqueMatchWithoutOrdinal(candidate) {
  switch (stringValue(candidate?.kind)) {
    case 'href':
    case 'role_label':
    case 'tag_label':
    case 'label':
    case 'placeholder':
    case 'tag_type':
    case 'tag':
    case 'type':
      return true;
    default:
      return false;
  }
}

function candidateSpecificityFields(candidate) {
  const kind = stringValue(candidate?.kind);
  const baseFields = new Set();
  switch (kind) {
    case 'role_label':
      baseFields.add('role');
      baseFields.add('label');
      break;
    case 'tag_label':
      baseFields.add('tag');
      baseFields.add('label');
      break;
    case 'label':
      baseFields.add('label');
      break;
    case 'placeholder':
      baseFields.add('placeholder');
      break;
    case 'tag_type':
      baseFields.add('tag');
      baseFields.add('type');
      break;
    case 'tag':
      baseFields.add('tag');
      break;
    case 'type':
      baseFields.add('tag');
      baseFields.add('type');
      break;
    case 'href':
      baseFields.add('href');
      break;
    default:
      break;
  }
  const out = [];
  const append = (field, value) => {
    if (!baseFields.has(field) && stringValue(value)) {
      out.push(field);
    }
  };
  append('role', candidate?.role);
  append('tag', candidate?.tag);
  append('label', candidate?.label);
  append('type', candidate?.type);
  append('href', candidate?.href);
  append('placeholder', candidate?.placeholder);
  return out;
}

function candidateStrength(candidate) {
  switch (stringValue(candidate?.kind)) {
    case 'href':
    case 'role_label':
    case 'tag_label':
      return 'strong';
    case 'label':
      return 'medium';
    case 'placeholder':
    case 'tag_type':
    case 'tag':
    case 'type':
      return 'weak';
    default:
      return '';
  }
}

function ambiguityClassForBlockedBy(blockedBy) {
  switch (stringValue(blockedBy)) {
    case 'multiple_candidates_same_semantic':
      return 'same_semantic';
    case 'multiple_candidates_filtered':
      return 'filtered_residual';
    case 'selector_index_out_of_range':
      return 'ordinal_out_of_range';
    case 'selector_index_filtered_out':
      return 'ordinal_filtered_out';
    default:
      return '';
  }
}

function retryDispositionForBlockedBy(blockedBy) {
  switch (stringValue(blockedBy)) {
    case 'multiple_candidates_same_semantic':
    case 'multiple_candidates_filtered':
    case 'selector_index_out_of_range':
    case 'selector_index_filtered_out':
      return 'manual_only';
    default:
      return '';
  }
}

function manualRetryHintForBlockedBy(blockedBy, recoveryAction = '') {
  switch (stringValue(blockedBy)) {
    case 'multiple_candidates':
    case 'multiple_candidates_same_semantic':
      return 'add_specificity';
    case 'multiple_candidates_filtered':
      return 'add_ordinal';
    case 'selector_index_out_of_range':
    case 'selector_index_filtered_out':
    case 'page_binding':
    case 'page_url':
    case 'page_origin':
    case 'page_path':
    case 'page_title':
    case 'tab_index':
    case 'frame_path':
      return 'refresh_snapshot';
    default:
      return nextStepAliasForRecoveryAction(recoveryAction) === 'snapshot' ? 'refresh_snapshot' : '';
  }
}

function nextStepAliasForRecoveryAction(recoveryAction) {
  const value = stringValue(recoveryAction);
  if (!value) {
    return '';
  }
  if (value === 'browser action=snapshot') {
    return 'snapshot';
  }
  if (value === 'browser action=refresh') {
    return 'refresh';
  }
  const match = /^browser action=([a-z_]+)$/i.exec(value);
  return match ? match[1].toLowerCase() : '';
}

function locatorCandidateScanOrder(count, preferredIndex, candidate) {
  if (count <= 0) {
    return [];
  }
  if (preferredIndex <= 0) {
    if (candidateRequiresUniqueMatchWithoutOrdinal(candidate)) {
      return count === 1 ? [0] : [];
    }
    return Array.from({ length: count }, (_, idx) => idx);
  }
  if (!candidateAllowsNearestOrdinalFallback(candidate)) {
    if (count === 1) {
      return [0];
    }
    if (preferredIndex >= 0 && preferredIndex < count) {
      return [preferredIndex];
    }
    return [];
  }
  const anchor = Math.max(0, Math.min(preferredIndex, count - 1));
  const order = [];
  const seen = new Set();
  const append = (value) => {
    if (value < 0 || value >= count || seen.has(value)) {
      return;
    }
    seen.add(value);
    order.push(value);
  };
  append(anchor);
  for (let offset = 1; order.length < count; offset += 1) {
    append(anchor - offset);
    append(anchor + offset);
  }
  return order;
}

function nearestMatchingIndexForPreferredOrdinal(matchingIndices, preferredIndex) {
  if (!Array.isArray(matchingIndices) || matchingIndices.length <= 0) {
    return -1;
  }
  const ordered = matchingIndices
    .filter((idx) => Number.isInteger(idx))
    .sort((left, right) => {
      const leftDistance = Math.abs(left - preferredIndex);
      const rightDistance = Math.abs(right - preferredIndex);
      if (leftDistance !== rightDistance) {
        return leftDistance - rightDistance;
      }
      return left - right;
    });
  return ordered.length > 0 ? ordered[0] : -1;
}

async function candidateWeakFallbackBlockedOutcome(locator, candidate) {
  if (!locator) {
    return null;
  }
  const count = await locator.count().catch(() => 0);
  if (count <= 0) {
    return null;
  }
  const preferredIndex = preferredOrdinalForLocatorCandidate(candidate);
  const matchingIndices = await locatorCandidateMatchingIndices(locator, candidate);
  if (preferredIndex <= 0) {
    if (candidateRequiresUniqueMatchWithoutOrdinal(candidate)) {
      if (matchingIndices.length > 1) {
        const kind = stringValue(candidate?.kind) || 'semantic';
        const blockedBy = matchingIndices.length < count ? 'multiple_candidates_filtered' : 'multiple_candidates_same_semantic';
        return {
          status: 'unresolved',
          candidate_kind: kind,
          candidate_strength: candidateStrength(candidate),
          ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
          retry_disposition: retryDispositionForBlockedBy(blockedBy),
          manual_retry_hint: manualRetryHintForBlockedBy(blockedBy, 'browser action=snapshot'),
          next_step_alias: 'snapshot',
          blocked_by: blockedBy,
          locator_count: count,
          candidate_count: matchingIndices.length,
          specificity_fields: candidateSpecificityFields(candidate),
          note: blockedBy === 'multiple_candidates_filtered'
            ? `${kind} fallback narrowed ${count} locator candidate(s) to ${matchingIndices.length} matching candidate(s), but a unique match is still required`
            : `${kind} fallback requires a unique same-semantic match but ${matchingIndices.length} matching candidate(s) remain`
        };
      }
      return null;
    }
    return null;
  }
  if (candidateAllowsNearestOrdinalFallback(candidate)) {
    return null;
  }
  if (matchingIndices.length <= 0 || matchingIndices.includes(preferredIndex)) {
    return null;
  }
  const kind = stringValue(candidate?.kind) || 'semantic';
  const blockedBy = preferredIndex >= count ? 'selector_index_out_of_range' : 'selector_index_filtered_out';
  return {
    status: 'unresolved',
    candidate_kind: kind,
    candidate_strength: candidateStrength(candidate),
    ambiguity_class: ambiguityClassForBlockedBy(blockedBy),
    retry_disposition: retryDispositionForBlockedBy(blockedBy),
    manual_retry_hint: manualRetryHintForBlockedBy(blockedBy, 'browser action=snapshot'),
    next_step_alias: 'snapshot',
    blocked_by: blockedBy,
    locator_count: count,
    candidate_count: matchingIndices.length,
    preferred_ordinal: preferredIndex + 1,
    specificity_fields: candidateSpecificityFields(candidate),
    note: blockedBy === 'selector_index_out_of_range'
      ? `${kind} fallback requires original ordinal ${preferredIndex + 1} but locator exposes only ${count} element(s)`
      : `${kind} fallback requires original ordinal ${preferredIndex + 1} but specificity filtered it out; ${matchingIndices.length} matching candidate(s) remain`
  };
}

async function locatorMatchesCandidate(locator, candidate) {
  const kind = stringValue(candidate?.kind);
  switch (kind) {
    case 'href':
    case 'role_label':
    case 'tag_label':
    case 'label':
    case 'placeholder':
    case 'tag_type':
    case 'tag':
    case 'type': {
      const expectedTag = candidateExpectedTag(candidate);
      const expectedType = stringValue(candidate?.type).toLowerCase();
      const expectedHref = stringValue(candidate?.href);
      const expectedPlaceholder = stringValue(candidate?.placeholder);
      const expectedLabel = stringValue(candidate?.label);
      try {
        const details = await locator.evaluate((el) => ({
          tag: String(el && el.tagName ? el.tagName : '').trim().toLowerCase(),
          type: String(el && el.getAttribute ? (el.getAttribute('type') || '') : '').trim().toLowerCase(),
          href: String(el && el.getAttribute ? (el.getAttribute('href') || '') : '').trim(),
          placeholder: String(el && el.getAttribute ? (el.getAttribute('placeholder') || '') : '').trim(),
          label: (() => {
            const root = el && el.ownerDocument ? el.ownerDocument : null;
            if (!el || !root) {
              return '';
            }
            const ariaLabelledBy = String(el.getAttribute ? (el.getAttribute('aria-labelledby') || '') : '').trim();
            if (ariaLabelledBy) {
              const pieces = [];
              ariaLabelledBy.split(/\s+/).forEach((id) => {
                const node = root.getElementById(id);
                const text = String(node && (node.textContent || node.innerText || '') || '').trim();
                if (text) {
                  pieces.push(text);
                }
              });
              if (pieces.length > 0) {
                return pieces.join(' ');
              }
            }
            const ariaLabel = String(el.getAttribute ? (el.getAttribute('aria-label') || '') : '').trim();
            if (ariaLabel) {
              return ariaLabel;
            }
            if (Array.isArray(el.labels) || (el.labels && typeof el.labels.length === 'number')) {
              const pieces = [];
              for (const labelEl of Array.from(el.labels || [])) {
                const text = String(labelEl && (labelEl.textContent || labelEl.innerText || '') || '').trim();
                if (text) {
                  pieces.push(text);
                }
              }
              if (pieces.length > 0) {
                return pieces.join(' ');
              }
            }
            const parentLabel = typeof el.closest === 'function' ? el.closest('label') : null;
            const parentLabelText = String(parentLabel && (parentLabel.textContent || parentLabel.innerText || '') || '').trim();
            if (parentLabelText) {
              return parentLabelText;
            }
            return String(
              (el.getAttribute ? (el.getAttribute('placeholder') || '') : '') ||
              (el.getAttribute ? (el.getAttribute('name') || '') : '') ||
              (el.getAttribute ? (el.getAttribute('title') || '') : '') ||
              el.textContent ||
              ''
            ).trim();
          })()
        }));
        if (expectedTag && stringValue(details?.tag) !== expectedTag) {
          return false;
        }
        if (expectedType && stringValue(details?.type).toLowerCase() !== expectedType) {
          return false;
        }
        if (expectedHref && stringValue(details?.href) !== expectedHref) {
          return false;
        }
        if (expectedPlaceholder && stringValue(details?.placeholder) !== expectedPlaceholder) {
          return false;
        }
        if (kind === 'href' && expectedLabel && stringValue(details?.label) !== expectedLabel) {
          return false;
        }
      } catch {
        return false;
      }
      return true;
    }
    default:
      return true;
  }
}

function candidateExpectedTag(candidate) {
  const explicitTag = stringValue(candidate?.tag).toLowerCase();
  if (explicitTag) {
    return explicitTag;
  }
  const inputType = stringValue(candidate?.type).toLowerCase();
  if (inputType) {
    return 'input';
  }
  return '';
}

function candidateAllowsFrameTraversal(candidate) {
  const kind = stringValue(candidate?.kind);
  switch (kind) {
    case 'native_ref':
    case 'page_binding':
      return false;
    default:
      return true;
  }
}

function descendantFramesForPage(page, preferredFramePath) {
  const entries = descendantFrameEntriesForPage(page);
  const preferred = stringValue(preferredFramePath);
  if (!preferred) {
    return entries.map((entry) => entry.frame);
  }
  const ranked = entries.map((entry, idx) => ({
    frame: entry.frame,
    path: entry.path,
    order: idx,
    score: framePathPreferenceScore(entry.path, preferred)
  }));
  sortRankedFrameEntries(ranked);
  return ranked.map((entry) => entry.frame);
}

function rankedDescendantFrameEntryGroupsForPage(page, preferredFramePath) {
  const preferred = stringValue(preferredFramePath);
  if (!preferred) {
    return [];
  }
  const ranked = descendantFrameEntriesForPage(page).map((entry, idx) => ({
    frame: entry.frame,
    path: entry.path,
    order: idx,
    score: framePathPreferenceScore(entry.path, preferred)
  }));
  sortRankedFrameEntries(ranked);
  const groups = [];
  for (const entry of ranked) {
    const last = groups[groups.length - 1];
    if (!last || !sameFramePreferenceScore(last[0].score, entry.score)) {
      groups.push([entry]);
      continue;
    }
    last.push(entry);
  }
  return groups;
}

function sortRankedFrameEntries(ranked) {
  ranked.sort((left, right) => {
    if (left.score.exact !== right.score.exact) {
      return left.score.exact ? -1 : 1;
    }
    if (left.score.anchorDepth !== right.score.anchorDepth) {
      return right.score.anchorDepth - left.score.anchorDepth;
    }
    if (left.score.tailAnchorDepth !== right.score.tailAnchorDepth) {
      return right.score.tailAnchorDepth - left.score.tailAnchorDepth;
    }
    if (left.score.tailSubsequenceDepth !== right.score.tailSubsequenceDepth) {
      return right.score.tailSubsequenceDepth - left.score.tailSubsequenceDepth;
    }
    if (left.score.tailGapCount !== right.score.tailGapCount) {
      return left.score.tailGapCount - right.score.tailGapCount;
    }
    if (left.score.commonPrefix !== right.score.commonPrefix) {
      return right.score.commonPrefix - left.score.commonPrefix;
    }
    if (left.score.commonSuffix !== right.score.commonSuffix) {
      return right.score.commonSuffix - left.score.commonSuffix;
    }
    if (left.score.distance !== right.score.distance) {
      return left.score.distance - right.score.distance;
    }
    if (left.score.divergence !== right.score.divergence) {
      return left.score.divergence - right.score.divergence;
    }
    return left.order - right.order;
  });
}

function sameFramePreferenceScore(left, right) {
  if (!left || !right) {
    return false;
  }
  if (
    !!left.exact !== !!right.exact ||
    intValue(left.anchorDepth) !== intValue(right.anchorDepth) ||
    intValue(left.tailAnchorDepth) !== intValue(right.tailAnchorDepth)
  ) {
    return false;
  }
  if (intValue(left.tailAnchorDepth) > 0) {
    // Preserve sort order inside a tail-anchored bucket, but treat tail-equivalent frames
    // as one preferred group so resolver fallback surfaces ambiguity instead of guessing.
    return intValue(left.commonSuffix) === intValue(right.commonSuffix);
  }
  if (intValue(left.tailSubsequenceDepth) > 0 || intValue(right.tailSubsequenceDepth) > 0) {
    return intValue(left.tailSubsequenceDepth) === intValue(right.tailSubsequenceDepth) &&
      intValue(left.tailGapCount) === intValue(right.tailGapCount);
  }
  return intValue(left.commonPrefix) === intValue(right.commonPrefix) &&
    intValue(left.commonSuffix) === intValue(right.commonSuffix) &&
    intValue(left.distance) === intValue(right.distance) &&
    intValue(left.divergence) === intValue(right.divergence);
}

function descendantFrameEntriesForPage(page) {
  const root = safeCall(() => page.mainFrame());
  if (!root || typeof root.childFrames !== 'function') {
    return [];
  }
  const out = [];
  const queue = Array.from(root.childFrames());
  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) {
      continue;
    }
    out.push({ frame: current, path: framePathForFrame(current, root) });
    for (const child of Array.from(current.childFrames())) {
      if (child) {
        queue.push(child);
      }
    }
  }
  return out;
}

function framePathPreferenceScore(currentPath, preferredPath) {
  const current = parseFramePath(currentPath);
  const preferred = parseFramePath(preferredPath);
  if (current.length === 0 || preferred.length === 0) {
    return {
      exact: false,
      anchorDepth: 0,
      tailAnchorDepth: 0,
      commonPrefix: 0,
      commonSuffix: 0,
      distance: Number.MAX_SAFE_INTEGER,
      divergence: Number.MAX_SAFE_INTEGER
    };
  }
  let commonPrefix = 0;
  while (commonPrefix < current.length && commonPrefix < preferred.length && current[commonPrefix] === preferred[commonPrefix]) {
    commonPrefix += 1;
  }
  let commonSuffix = 0;
  while (
    commonSuffix < current.length &&
    commonSuffix < preferred.length &&
    current[current.length - 1 - commonSuffix] === preferred[preferred.length - 1 - commonSuffix]
  ) {
    commonSuffix += 1;
  }
  const anchoredTailDepth = commonSuffix === preferred.length
    ? preferred.length
    : commonSuffix === current.length
      ? current.length
      : 0;
  const tailSubsequence = tailSubsequencePreferenceScore(current, preferred);
  const tailAnchorDepth = anchoredTailDepth > 0 || commonSuffix >= 2 ? commonSuffix : 0;
  const divergence = commonPrefix < current.length && commonPrefix < preferred.length
    ? Math.abs(current[commonPrefix] - preferred[commonPrefix])
    : 0;
  return {
    exact: stringValue(currentPath) === stringValue(preferredPath),
    // Preserve subtree rebinding preference when the original frame tail survives under a new parent chain.
    anchorDepth: Math.max(commonPrefix, tailAnchorDepth, tailSubsequence.depth),
    tailAnchorDepth,
    tailSubsequenceDepth: tailSubsequence.depth,
    tailGapCount: tailSubsequence.gapCount,
    commonPrefix,
    commonSuffix,
    distance: (current.length - commonPrefix) + (preferred.length - commonPrefix),
    divergence
  };
}

function tailSubsequencePreferenceScore(current, preferred) {
  if (!Array.isArray(current) || !Array.isArray(preferred) || current.length <= 0 || preferred.length <= 0) {
    return {
      depth: 0,
      gapCount: Number.MAX_SAFE_INTEGER
    };
  }
  let currentIndex = current.length - 1;
  let preferredIndex = preferred.length - 1;
  if (current[currentIndex] !== preferred[preferredIndex]) {
    return {
      depth: 0,
      gapCount: Number.MAX_SAFE_INTEGER
    };
  }
  let depth = 0;
  let gapCount = 0;
  while (currentIndex >= 0 && preferredIndex >= 0) {
    if (current[currentIndex] === preferred[preferredIndex]) {
      depth += 1;
      currentIndex -= 1;
      preferredIndex -= 1;
      continue;
    }
    gapCount += 1;
    currentIndex -= 1;
  }
  if (depth < 2) {
    return {
      depth: 0,
      gapCount: Number.MAX_SAFE_INTEGER
    };
  }
  return {
    depth,
    gapCount
  };
}

function parseFramePath(value) {
  const normalized = stringValue(value);
  if (!normalized) {
    return [];
  }
  const parts = [];
  for (const rawPart of normalized.split('/')) {
    const trimmed = stringValue(rawPart);
    if (!/^\d+$/.test(trimmed)) {
      return [];
    }
    parts.push(Number.parseInt(trimmed, 10));
  }
  return parts;
}

function framePathForFrame(frame, root) {
  if (!frame || !root || frame === root) {
    return '';
  }
  const parts = [];
  let current = frame;
  while (current && current !== root) {
    const parent = safeCall(() => current.parentFrame());
    if (!parent) {
      return '';
    }
    const siblings = Array.from(parent.childFrames());
    const index = siblings.indexOf(current);
    if (index < 0) {
      return '';
    }
    parts.push(String(index));
    current = parent;
  }
  return parts.reverse().join('/');
}

function locatorForCandidate(page, candidate) {
  const kind = stringValue(candidate.kind);
  switch (kind) {
    case 'selector':
      return { locator: locatorForCandidateInScope(page, candidate), outcome: null };
    case 'native_ref': {
      const nativeRef = stringValue(candidate.native_ref);
      const registered = locatorForNativeRef(page, nativeRef);
      if (registered.locator) {
        return registered;
      }
      const selector = selectorFromElementRef(nativeRef);
      return {
        locator: selector ? page.locator(selector) : null,
        outcome: registered.outcome
      };
    }
    default:
      return { locator: locatorForCandidateInScope(page, candidate), outcome: null };
  }
}

function locatorForCandidateInScope(scope, candidate) {
  const kind = stringValue(candidate.kind);
  switch (kind) {
    case 'href':
      return stringValue(candidate.href) ? scope.locator(`a[href=${JSON.stringify(stringValue(candidate.href))}]`) : null;
    case 'role_label':
      if (!stringValue(candidate.role) || !stringValue(candidate.label)) {
        return null;
      }
      return scope.getByRole(stringValue(candidate.role), { name: stringValue(candidate.label), exact: true });
    case 'tag_label':
      if (!stringValue(candidate.label)) {
        return null;
      }
      return locatorForTagLabelCandidate(scope, candidate);
    case 'label':
      return stringValue(candidate.label) ? scope.getByLabel(stringValue(candidate.label), { exact: true }) : null;
    case 'placeholder':
      return stringValue(candidate.placeholder) ? scope.getByPlaceholder(stringValue(candidate.placeholder), { exact: true }) : null;
    case 'tag_type':
      if (!stringValue(candidate.tag) || !stringValue(candidate.type)) {
        return null;
      }
      return scope.locator(`${stringValue(candidate.tag)}[type=${JSON.stringify(stringValue(candidate.type))}]`);
    case 'tag':
      return stringValue(candidate.tag) ? scope.locator(stringValue(candidate.tag)) : null;
    case 'type':
      if (!stringValue(candidate.type)) {
        return null;
      }
      {
        const tag = candidateExpectedTag(candidate);
        return tag
          ? scope.locator(`${tag}[type=${JSON.stringify(stringValue(candidate.type))}]`)
          : scope.locator(`[type=${JSON.stringify(stringValue(candidate.type))}]`);
      }
    case 'selector':
      if (!stringValue(candidate.selector)) {
        return null;
      }
      {
        const locator = scope.locator(stringValue(candidate.selector));
        return intValue(candidate.selector_index) > 0 ? locator.nth(intValue(candidate.selector_index)) : locator;
      }
    default:
      return null;
  }
}

function locatorForTagLabelCandidate(scope, candidate) {
  const tag = candidateExpectedTag(candidate);
  const label = stringValue(candidate.label);
  if (!label) {
    return null;
  }
  switch (tag) {
    case 'input':
    case 'textarea':
    case 'select':
      return scope.getByLabel(label, { exact: true });
    default:
      return tag ? scope.locator(tag).filter({ hasText: label }) : scope.getByLabel(label, { exact: true });
  }
}

function locatorCandidatesFromDescriptor(descriptor) {
  const source = normalizeRequestParams(descriptor);
  const out = [];
  if (stringValue(source.native_ref)) {
    appendUniqueLocatorCandidate(out, { kind: 'native_ref', native_ref: stringValue(source.native_ref) });
  }
  if (stringValue(source.selector)) {
    appendUniqueLocatorCandidate(out, { kind: 'selector', selector: stringValue(source.selector), selector_index: intValue(source.selector_index) });
  }
  if (stringValue(source.href)) {
    appendUniqueLocatorCandidate(out, { kind: 'href', href: stringValue(source.href), selector_index: intValue(source.selector_index) });
  }
  if (stringValue(source.role) && stringValue(source.label)) {
    appendUniqueLocatorCandidate(out, {
      kind: 'role_label',
      role: stringValue(source.role),
      tag: stringValue(source.tag),
      label: stringValue(source.label),
      type: stringValue(source.type),
      href: stringValue(source.href),
      placeholder: stringValue(source.placeholder),
      selector_index: intValue(source.selector_index)
    });
  } else if (stringValue(source.tag) && stringValue(source.label)) {
    appendUniqueLocatorCandidate(out, {
      kind: 'tag_label',
      tag: stringValue(source.tag),
      type: stringValue(source.type),
      label: stringValue(source.label),
      href: stringValue(source.href),
      placeholder: stringValue(source.placeholder),
      selector_index: intValue(source.selector_index)
    });
  } else if (stringValue(source.label)) {
    appendUniqueLocatorCandidate(out, {
      kind: 'label',
      tag: stringValue(source.tag),
      label: stringValue(source.label),
      type: stringValue(source.type),
      href: stringValue(source.href),
      placeholder: stringValue(source.placeholder),
      selector_index: intValue(source.selector_index)
    });
  }
  if (stringValue(source.placeholder)) {
    appendUniqueLocatorCandidate(out, {
      kind: 'placeholder',
      tag: stringValue(source.tag),
      type: stringValue(source.type),
      placeholder: stringValue(source.placeholder),
      selector_index: intValue(source.selector_index)
    });
  }
  if (stringValue(source.tag) && stringValue(source.type)) {
    appendUniqueLocatorCandidate(out, { kind: 'tag_type', tag: stringValue(source.tag), type: stringValue(source.type), selector_index: intValue(source.selector_index) });
  } else {
    if (stringValue(source.tag)) {
      appendUniqueLocatorCandidate(out, { kind: 'tag', tag: stringValue(source.tag), selector_index: intValue(source.selector_index) });
    }
    if (stringValue(source.type)) {
      appendUniqueLocatorCandidate(out, { kind: 'type', type: stringValue(source.type), selector_index: intValue(source.selector_index) });
    }
  }
  return out;
}

function pageBindingFromDescriptor(descriptor) {
  const source = normalizeRequestParams(descriptor);
  const candidate = normalizePageBindingCandidate({
    kind: 'page_binding',
    page_url: stringValue(source.page_url),
    page_origin: stringValue(source.page_origin),
    page_path: stringValue(source.page_path),
    page_title: stringValue(source.page_title),
    tab_index: intValue(source.tab_index)
  });
  return candidate;
}

function normalizePageBindingCandidate(candidate) {
  const current = normalizeRequestParams(candidate);
  if (!stringValue(current.page_url) &&
    !stringValue(current.page_origin) &&
    !stringValue(current.page_path) &&
    !stringValue(current.page_title) &&
    intValue(current.tab_index) <= 0 &&
    !stringValue(current.frame_path)) {
    return null;
  }
  return {
    kind: 'page_binding',
    page_url: stringValue(current.page_url),
    page_origin: stringValue(current.page_origin),
    page_path: stringValue(current.page_path),
    page_title: stringValue(current.page_title),
    tab_index: intValue(current.tab_index),
    frame_path: stringValue(current.frame_path)
  };
}

function appendUniqueLocatorCandidate(target, candidate) {
  const normalized = normalizeLocatorCandidate(candidate);
  if (!locatorCandidateValid(normalized)) {
    return;
  }
  const key = locatorCandidateKey(normalized);
  if (!key) {
    return;
  }
  const existingIndex = target.findIndex((item) => locatorCandidateKey(item) === key);
  if (existingIndex >= 0) {
    target[existingIndex] = mergeLocatorCandidate(target[existingIndex], normalized);
    return;
  }
  target.push(normalized);
}

function mergeLocatorCandidate(base, extra) {
  const current = normalizeLocatorCandidate(base);
  const incoming = normalizeLocatorCandidate(extra);
  return {
    kind: firstNonEmpty(stringValue(current.kind), stringValue(incoming.kind)),
    selector: firstNonEmpty(stringValue(current.selector), stringValue(incoming.selector)),
    selector_index: intValue(current.selector_index) > 0 ? intValue(current.selector_index) : intValue(incoming.selector_index),
    frame_path: firstNonEmpty(stringValue(current.frame_path), stringValue(incoming.frame_path)),
    native_ref: firstNonEmpty(stringValue(current.native_ref), stringValue(incoming.native_ref)),
    role: firstNonEmpty(stringValue(current.role), stringValue(incoming.role)),
    tag: firstNonEmpty(stringValue(current.tag), stringValue(incoming.tag)),
    label: firstNonEmpty(stringValue(current.label), stringValue(incoming.label)),
    type: firstNonEmpty(stringValue(current.type), stringValue(incoming.type)),
    href: firstNonEmpty(stringValue(current.href), stringValue(incoming.href)),
    placeholder: firstNonEmpty(stringValue(current.placeholder), stringValue(incoming.placeholder)),
    page_url: firstNonEmpty(stringValue(current.page_url), stringValue(incoming.page_url)),
    page_origin: firstNonEmpty(stringValue(current.page_origin), stringValue(incoming.page_origin)),
    page_path: firstNonEmpty(stringValue(current.page_path), stringValue(incoming.page_path)),
    page_title: firstNonEmpty(stringValue(current.page_title), stringValue(incoming.page_title)),
    tab_index: intValue(current.tab_index) > 0 ? intValue(current.tab_index) : intValue(incoming.tab_index)
  };
}

function normalizeLocatorCandidate(candidate) {
  const current = normalizeRequestParams(candidate);
  return {
    kind: stringValue(current.kind),
    selector: stringValue(current.selector),
    selector_index: intValue(current.selector_index),
    frame_path: stringValue(current.frame_path),
    native_ref: stringValue(current.native_ref),
    role: stringValue(current.role),
    tag: stringValue(current.tag),
    label: stringValue(current.label),
    type: stringValue(current.type),
    href: stringValue(current.href),
    placeholder: stringValue(current.placeholder),
    page_url: stringValue(current.page_url),
    page_origin: stringValue(current.page_origin),
    page_path: stringValue(current.page_path),
    page_title: stringValue(current.page_title),
    tab_index: intValue(current.tab_index)
  };
}

function locatorCandidateValid(candidate) {
  switch (stringValue(candidate.kind)) {
    case 'native_ref':
      return stringValue(candidate.native_ref) !== '';
    case 'selector':
      return stringValue(candidate.selector) !== '';
    case 'href':
      return stringValue(candidate.href) !== '';
    case 'role_label':
      return stringValue(candidate.role) !== '' && stringValue(candidate.label) !== '';
    case 'tag_label':
      return stringValue(candidate.tag) !== '' && stringValue(candidate.label) !== '';
    case 'label':
      return stringValue(candidate.label) !== '';
    case 'placeholder':
      return stringValue(candidate.placeholder) !== '';
    case 'tag_type':
      return stringValue(candidate.tag) !== '' && stringValue(candidate.type) !== '';
    case 'tag':
      return stringValue(candidate.tag) !== '';
    case 'type':
      return stringValue(candidate.type) !== '';
    case 'page_binding':
      return normalizePageBindingCandidate(candidate) !== null;
    default:
      return false;
  }
}

function locatorCandidateKey(candidate) {
  const current = normalizeLocatorCandidate(candidate);
  switch (stringValue(current.kind)) {
    case 'native_ref':
      return `native_ref|${current.native_ref}|${current.frame_path}`;
    case 'selector':
      return `selector|${current.selector}|${current.selector_index}|${current.frame_path}`;
    case 'href':
      return `href|${current.href}|${current.selector_index}|${current.frame_path}`;
    case 'role_label':
      return `role_label|${current.role}|${current.label}|${current.tag}|${current.type}|${current.href}|${current.placeholder}|${current.selector_index}|${current.frame_path}`;
    case 'tag_label':
      return `tag_label|${current.tag}|${current.type}|${current.label}|${current.href}|${current.placeholder}|${current.selector_index}|${current.frame_path}`;
    case 'label':
      return `label|${current.label}|${current.tag}|${current.type}|${current.href}|${current.placeholder}|${current.selector_index}|${current.frame_path}`;
    case 'placeholder':
      return `placeholder|${current.placeholder}|${current.tag}|${current.type}|${current.selector_index}|${current.frame_path}`;
    case 'tag_type':
      return `tag_type|${current.tag}|${current.type}|${current.selector_index}|${current.frame_path}`;
    case 'tag':
      return `tag|${current.tag}|${current.selector_index}|${current.frame_path}`;
    case 'type':
      return `type|${current.tag}|${current.type}|${current.selector_index}|${current.frame_path}`;
    case 'page_binding':
      return `page_binding|${current.page_url}|${current.page_origin}|${current.page_path}|${current.page_title}|${current.tab_index}|${current.frame_path}`;
    default:
      return '';
  }
}

function resolverMatchedNote(candidate) {
  const kind = stringValue(candidate.kind);
  return kind ? `resolved via ${kind}` : '';
}

function normalizeResolverOutcome(outcome) {
  if (!outcome || typeof outcome !== 'object') {
    return null;
  }
  const normalized = {
    status: stringValue(outcome.status),
    resolution_mode: stringValue(outcome.resolution_mode),
    primary_kind: stringValue(outcome.primary_kind),
    attempt_count: intValue(outcome.attempt_count),
    matched_kind: stringValue(outcome.matched_kind),
    matched_index: intValue(outcome.matched_index),
    matched_candidate_kind: stringValue(outcome.matched_candidate_kind),
    fallback_from_kind: stringValue(outcome.fallback_from_kind),
    fallback_from_index: intValue(outcome.fallback_from_index),
    fallback_from_blocked_by: stringValue(outcome.fallback_from_blocked_by),
    fallback_from_ambiguity_class: stringValue(outcome.fallback_from_ambiguity_class),
    fallback_from_candidate_strength: stringValue(outcome.fallback_from_candidate_strength),
    fallback_from_manual_retry_hint: stringValue(outcome.fallback_from_manual_retry_hint),
    fallback_from_specificity_fields: stringSliceValue(outcome.fallback_from_specificity_fields),
    candidate_kind: stringValue(outcome.candidate_kind),
    candidate_strength: stringValue(outcome.candidate_strength),
    ambiguity_class: stringValue(outcome.ambiguity_class),
    retry_disposition: stringValue(outcome.retry_disposition),
    manual_retry_hint: stringValue(outcome.manual_retry_hint),
    next_step_alias: stringValue(outcome.next_step_alias),
    blocked_by: stringValue(outcome.blocked_by),
    locator_count: intValue(outcome.locator_count),
    candidate_count: intValue(outcome.candidate_count),
    preferred_ordinal: intValue(outcome.preferred_ordinal),
    specificity_fields: stringSliceValue(outcome.specificity_fields),
    recovery_action: stringValue(outcome.recovery_action),
    note: stringValue(outcome.note)
  };
  const resolvedFramePath = stringValue(outcome.resolved_frame_path);
  if (resolvedFramePath) {
    normalized.resolved_frame_path = resolvedFramePath;
  }
  const resolvedSelectorIndex = intValue(outcome.resolved_selector_index);
  if (resolvedSelectorIndex > 0) {
    normalized.resolved_selector_index = resolvedSelectorIndex;
  }
  if (outcome.native_ref_rebound === true) {
    normalized.native_ref_rebound = true;
  }
  if (!normalized.matched_candidate_kind && normalized.status === 'matched') {
    normalized.matched_candidate_kind = normalized.matched_kind;
  }
  if (!normalized.recovery_action) {
    switch (normalized.status) {
      case 'page_binding_blocked':
      case 'unresolved':
        normalized.recovery_action = 'browser action=snapshot';
        break;
      case 'resolution_failed':
        normalized.recovery_action = ['page_binding', 'page_url', 'page_origin', 'page_path', 'page_title', 'tab_index', 'frame_path'].includes(normalized.blocked_by)
          ? 'browser action=snapshot'
          : 'browser action=refresh';
        break;
      default:
        break;
    }
  }
  if (!normalized.next_step_alias) {
    normalized.next_step_alias = nextStepAliasForRecoveryAction(normalized.recovery_action);
  }
  if (!normalized.manual_retry_hint) {
    normalized.manual_retry_hint = manualRetryHintForBlockedBy(normalized.blocked_by, normalized.recovery_action);
  }
  return normalized;
}

async function browserActionabilityReportForLocator(action, locator, resolverOutcome, params = {}) {
  const normalizedAction = normalizeActionabilityAction(action);
  if (!normalizedAction || !locator) {
    return null;
  }
  const outcome = normalizeResolverOutcome(resolverOutcome);
  const report = {
    action: normalizedAction,
    status: 'partial',
    target_kind: actionabilityTargetKind(params),
    target: actionabilityTargetValue(params),
    checks: []
  };

  let resolutionFailed = false;
  const resolveCheck = {
    name: 'resolve_target',
    status: 'not_reported',
    required: true
  };
  if (outcome) {
    resolveCheck.detail = stringValue(outcome.note);
    if (stringValue(outcome.status) === 'matched') {
      resolveCheck.status = 'passed';
    } else if (stringValue(outcome.status)) {
      resolveCheck.status = 'failed';
      resolutionFailed = true;
      report.status = 'failed';
      report.failed_check = 'resolve_target';
      report.failure_reason = actionabilityFailureReason(outcome);
      report.retry_disposition = stringValue(outcome.retry_disposition);
      report.manual_retry_hint = stringValue(outcome.manual_retry_hint);
      report.recovery_action = stringValue(outcome.recovery_action);
    }
  }
  report.checks.push(resolveCheck);

  for (const name of actionabilityRequiredChecks(normalizedAction)) {
    if (resolutionFailed) {
      report.checks.push({
        name,
        status: 'skipped',
        required: true,
        detail: 'blocked by target resolution'
      });
      continue;
    }
    const check = await evaluateActionabilityCheck(locator, name, params);
    report.checks.push(check);
    if (check.required && check.status === 'failed' && report.status !== 'failed') {
      report.status = 'failed';
      report.failed_check = name;
      report.failure_reason = `actionability_${name}_failed`;
      report.manual_retry_hint = manualRetryHintForActionabilityCheck(name);
      report.recovery_action = recoveryActionForActionabilityCheck(name);
    }
  }
  if (report.status !== 'failed' && report.checks.length > 0 && report.checks.every((check) => !check.required || check.status === 'passed')) {
    report.status = 'passed';
  }
  return report;
}

function normalizeActionabilityAction(action) {
  const value = stringValue(action).toLowerCase();
  return value === 'type_text' ? 'type' : value;
}

function applyActionabilityCheckOutcome(report, name, outcome) {
  if (!report || !Array.isArray(report.checks)) {
    return report;
  }
  const normalizedName = stringValue(name);
  if (!normalizedName) {
    return report;
  }
  const check = report.checks.find((item) => stringValue(item?.name) === normalizedName);
  if (!check) {
    return report;
  }
  const normalizedOutcome = normalizeActionabilityCheckOutcome(outcome);
  if (!normalizedOutcome.status) {
    return report;
  }
  check.status = normalizedOutcome.status;
  check.detail = normalizedOutcome.detail;
  if (check.required && check.status === 'failed') {
    if (report.status !== 'failed' || !report.failed_check) {
      report.status = 'failed';
      report.failed_check = normalizedName;
      report.failure_reason = `actionability_${normalizedName}_failed`;
      report.manual_retry_hint = manualRetryHintForActionabilityCheck(normalizedName);
      report.recovery_action = recoveryActionForActionabilityCheck(normalizedName);
    }
    return report;
  }
  if (report.status !== 'failed') {
    report.status = report.checks.every((item) => !item.required || item.status === 'passed') ? 'passed' : 'partial';
  }
  return report;
}

function normalizeActionabilityCheckOutcome(outcome) {
  if (typeof outcome === 'string') {
    return { status: stringValue(outcome), detail: '' };
  }
  if (!outcome || typeof outcome !== 'object') {
    return { status: '', detail: '' };
  }
  return {
    status: stringValue(outcome.status),
    detail: stringValue(outcome.detail)
  };
}

function actionabilityRequiredChecks(action) {
  switch (normalizeActionabilityAction(action)) {
    case 'click':
      return ['attached', 'visible', 'stable', 'receives_events', 'enabled', 'frame_hit_target', 'navigation_wait'];
    case 'type':
    case 'fill':
      return ['attached', 'visible', 'stable', 'enabled', 'editable'];
    case 'select':
      return ['attached', 'visible', 'stable', 'enabled'];
    case 'upload':
      return ['attached', 'enabled'];
    case 'hover':
    case 'drag':
      return ['attached', 'visible', 'stable', 'receives_events', 'frame_hit_target'];
    case 'screenshot':
    case 'highlight':
      return ['attached', 'visible', 'stable'];
    default:
      return [];
  }
}

async function evaluateActionabilityCheck(locator, name, params = {}) {
  const check = { name, status: 'not_reported', required: true };
  try {
    switch (name) {
      case 'attached': {
        const count = await locator.count();
        check.status = count > 0 ? 'passed' : 'failed';
        if (count > 1) {
          check.detail = `locator matched ${count} elements`;
        }
        return check;
      }
      case 'visible':
        check.status = await locator.isVisible() ? 'passed' : 'failed';
        return check;
      case 'enabled':
        check.status = await locator.isEnabled() ? 'passed' : 'failed';
        return check;
      case 'editable':
        check.status = await locator.isEditable() ? 'passed' : 'failed';
        return check;
      case 'stable':
        check.status = await locator.evaluate(actionabilityStableEvaluate) ? 'passed' : 'failed';
        return check;
      case 'receives_events':
        check.status = await locator.evaluate(actionabilityReceivesEventsEvaluate) ? 'passed' : 'failed';
        return check;
      case 'frame_hit_target': {
        const outcome = await actionabilityFrameHitTargetOutcome(locator);
        check.status = outcome.passed ? 'passed' : 'failed';
        check.detail = outcome.detail;
        return check;
      }
      case 'navigation_wait':
        check.status = 'not_reported';
        check.detail = 'navigation completion is observed by the action execution path';
        return check;
      default:
        check.status = 'not_reported';
        return check;
    }
  } catch (err) {
    check.status = 'failed';
    check.detail = errorMessage(err);
    return check;
  }
}

async function actionabilityStableEvaluate(element) {
  if (!(element instanceof Element)) {
    return false;
  }
  const first = element.getBoundingClientRect();
  await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)));
  const second = element.getBoundingClientRect();
  return first.x === second.x &&
    first.y === second.y &&
    first.width === second.width &&
    first.height === second.height &&
    first.width > 0 &&
    first.height > 0;
}

function actionabilityReceivesEventsEvaluate(element) {
  if (!(element instanceof Element)) {
    return false;
  }
  const rect = element.getBoundingClientRect();
  if (!rect || rect.width <= 0 || rect.height <= 0) {
    return false;
  }
  const x = Math.min(Math.max(rect.left + rect.width / 2, 0), Math.max(document.documentElement.clientWidth - 1, 0));
  const y = Math.min(Math.max(rect.top + rect.height / 2, 0), Math.max(document.documentElement.clientHeight - 1, 0));
  const hit = document.elementFromPoint(x, y);
  return hit === element || Boolean(hit && element.contains(hit));
}

async function actionabilityFrameHitTargetOutcome(locator) {
  const handle = await locator.elementHandle();
  if (!handle) {
    return { passed: false, detail: 'target element handle unavailable' };
  }
  try {
    const local = await handle.evaluate(actionabilityHitTargetDetailEvaluate);
    if (!local?.passed) {
      return {
        passed: false,
        detail: `target_frame ${stringValue(local?.detail) || 'center point is not hittable'}`
      };
    }

    let frame = await handle.ownerFrame();
    let depth = 0;
    while (frame && frame.parentFrame()) {
      depth += 1;
      let frameElement;
      try {
        frameElement = await frame.frameElement();
      } catch (err) {
        return {
          passed: false,
          detail: `parent_frame_depth=${depth} frame_element_unavailable ${errorMessage(err)}`
        };
      }
      try {
        const parentHit = await frameElement.evaluate(actionabilityHitTargetDetailEvaluate);
        if (!parentHit?.passed) {
          return {
            passed: false,
            detail: `parent_frame_depth=${depth} ${stringValue(parentHit?.detail) || 'frame element center point is not hittable'}`
          };
        }
      } finally {
        await frameElement.dispose().catch(() => {});
      }
      frame = frame.parentFrame();
    }
    return {
      passed: true,
      detail: depth > 0 ? `parent_frame_chain_passed depth=${depth}` : 'top_frame'
    };
  } finally {
    await handle.dispose().catch(() => {});
  }
}

function actionabilityHitTargetDetailEvaluate(element) {
  if (!(element instanceof Element)) {
    return { passed: false, detail: 'target is not an element' };
  }
  const rect = element.getBoundingClientRect();
  if (!rect || rect.width <= 0 || rect.height <= 0) {
    return { passed: false, detail: 'target has no visible bounding box' };
  }
  const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
  const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
  if (viewportWidth <= 0 || viewportHeight <= 0) {
    return { passed: false, detail: 'viewport has no hittable area' };
  }
  const x = Math.min(Math.max(rect.left + rect.width / 2, 0), viewportWidth - 1);
  const y = Math.min(Math.max(rect.top + rect.height / 2, 0), viewportHeight - 1);
  const hit = document.elementFromPoint(x, y);
  const passed = hit === element || Boolean(hit && element.contains(hit));
  if (passed) {
    return { passed: true, detail: `center=${Math.round(x)},${Math.round(y)}` };
  }
  const hitLabel = hit instanceof Element
    ? `${hit.tagName.toLowerCase()}${hit.id ? `#${hit.id}` : ''}${typeof hit.className === 'string' && hit.className ? `.${hit.className.trim().replace(/\s+/g, '.')}` : ''}`
    : 'none';
  return { passed: false, detail: `center=${Math.round(x)},${Math.round(y)} hit=${hitLabel}` };
}

function actionabilityTargetKind(params = {}) {
  if (stringValue(params.element_ref) || stringValue(params.ref)) {
    return 'ref';
  }
  if (stringValue(params.selector)) {
    return 'selector';
  }
  return '';
}

function actionabilityTargetValue(params = {}) {
  return firstNonEmpty(params.element_ref, params.ref, params.selector);
}

function actionabilityFailureReason(outcome) {
  const status = stringValue(outcome?.status);
  const blockedBy = stringValue(outcome?.blocked_by);
  if (status && blockedBy) {
    return `resolver_${status}_${blockedBy}`;
  }
  return status ? `resolver_${status}` : '';
}

function actionabilityFailedOnAny(actionability, names = []) {
  if (stringValue(actionability?.status) !== 'failed') {
    return false;
  }
  const failedCheck = stringValue(actionability?.failed_check);
  return arrayValue(names).map((name) => stringValue(name)).includes(failedCheck);
}

function fieldActionabilityFailureMessage(actionability) {
  return firstNonEmpty(
    stringValue(actionability?.failure_reason),
    stringValue(actionability?.failed_check) ? `actionability_${stringValue(actionability.failed_check)}_failed` : '',
    'actionability_failed'
  );
}

function manualRetryHintForActionabilityCheck(name) {
  switch (name) {
    case 'attached':
      return 'refresh_snapshot';
    case 'visible':
    case 'stable':
    case 'receives_events':
    case 'frame_hit_target':
      return 'wait_or_choose_visible_target';
    case 'navigation_wait':
      return 'wait_or_refresh_snapshot';
    case 'enabled':
    case 'editable':
      return 'choose_enabled_editable_target';
    default:
      return '';
  }
}

function recoveryActionForActionabilityCheck(name) {
  switch (name) {
    case 'attached':
      return 'browser action=snapshot';
    case 'visible':
    case 'stable':
    case 'receives_events':
    case 'frame_hit_target':
    case 'enabled':
    case 'editable':
    case 'navigation_wait':
      return 'browser action=snapshot';
    default:
      return '';
  }
}

function actionFailureReasonCode(actionability) {
  if (stringValue(actionability?.status) === 'failed' && stringValue(actionability?.failure_reason)) {
    return stringValue(actionability.failure_reason);
  }
  return 'action_failed';
}

function compactObject(value) {
  const out = {};
  for (const [key, current] of Object.entries(value || {})) {
    if (current === null || current === undefined) {
      continue;
    }
    if (typeof current === 'string' && current === '') {
      continue;
    }
    if (Array.isArray(current) && current.length === 0) {
      continue;
    }
    if (typeof current === 'number' && current === 0) {
      continue;
    }
    if (typeof current === 'boolean' && current === false) {
      continue;
    }
    out[key] = current;
  }
  return out;
}

function browserActionFailureArtifact(action, page, target, actionability, reasonCode, note, extra = {}) {
  const resolverOutcome = normalizeResolverOutcome(target?.outcome);
  const finalURL = page && typeof page.url === 'function' ? page.url() : '';
  const state = page ? pageSessionStateForPage(page) : null;
  const errors = state ? browserErrorEntriesForState(state) : [];
  const artifactPathValue = firstNonEmpty(extra.artifact_path, extra.artifactPath, extra.path);
  const screenshotPath = firstNonEmpty(extra.screenshot_path, extra.screenshotPath);
  const normalizedAction = normalizeActionabilityAction(action);
  const artifact = {
    kind: 'trace_like',
    action: normalizedAction,
    status: 'action_failed',
    reason_code: stringValue(reasonCode),
    message: stringValue(note),
    recovery_action: firstNonEmpty(actionability?.recovery_action, resolverOutcome?.recovery_action, 'browser action=snapshot'),
    final_url: finalURL,
    title: stringValue(extra.title),
    target_kind: stringValue(actionability?.target_kind),
    target: stringValue(actionability?.target),
    failed_check: stringValue(actionability?.failed_check),
    snapshot_available: Boolean(extra.snapshot_available),
    snapshot_format: stringValue(extra.snapshot_format),
    snapshot_refs: stringValue(extra.snapshot_refs),
    snapshot_frame: stringValue(extra.snapshot_frame),
    snapshot_element_count: intValue(extra.snapshot_element_count),
    snapshot_truncated: Boolean(extra.snapshot_truncated),
    screenshot_path: screenshotPath,
    artifact_path: artifactPathValue,
    resolver_outcome: resolverOutcome,
    console_message_count: 0,
    error_count: errors.length,
    errors
  };
  return compactObject(artifact);
}

function browserActionFailureResult(action, page, target, actionability, err, extra = {}) {
  const resolverOutcome = normalizeResolverOutcome(target?.outcome);
  const note = firstNonEmpty(extra.note, errorMessage(err));
  const finalURL = page && typeof page.url === 'function' ? page.url() : '';
  const reasonCode = actionFailureReasonCode(actionability);
  const artifactPathValue = firstNonEmpty(extra.artifact_path, extra.artifactPath, extra.path);
  const artifact = browserActionFailureArtifact(action, page, target, actionability, reasonCode, note, extra);
  return {
    backend: backendName,
    browser_app: browserApp,
    final_url: finalURL,
    title: '',
    status: 'action_failed',
    note,
    ...extra,
    actionability,
    failure_evidence: {
      action: normalizeActionabilityAction(action),
      status: 'action_failed',
      reason_code: reasonCode,
      message: note,
      retryable: true,
      recovery_action: firstNonEmpty(actionability?.recovery_action, resolverOutcome?.recovery_action, 'browser action=snapshot'),
      resolver_outcome: resolverOutcome,
      actionability,
      artifact_path: artifactPathValue,
      artifact
    },
    resolver_outcome: resolverOutcome
  };
}

function resolverHttpError(statusCode, message, outcome) {
  const err = httpError(statusCode, message);
  err.resolverOutcome = normalizeResolverOutcome(outcome);
  return err;
}

function errorResolverOutcome(err) {
  return normalizeResolverOutcome(err && err.resolverOutcome);
}

function actionTargetRequested(params) {
  return stringValue(params.selector) !== '' ||
    stringValue(params.element_ref) !== '' ||
    Object.keys(objectValue(params.element_hint)).length > 0 ||
    Object.keys(objectValue(params.element_resolver)).length > 0;
}

function uploadTargetRequested(params) {
  return actionTargetRequested(uploadActionLocatorParams(params));
}

function fillFieldActionLocatorParams(field) {
  const current = normalizeRequestParams(field);
  return {
    selector: stringValue(current.selector),
    element_ref: firstNonEmpty(
      stringValue(current.ref),
      stringValue(current.element_ref)
    ),
    element_hint: objectValue(current.hint),
    element_resolver: objectValue(current.resolver)
  };
}

function uploadActionLocatorParams(params) {
  const current = normalizeRequestParams(params);
  current.element_ref = firstNonEmpty(
    stringValue(current.input_ref),
    stringValue(current.ref),
    stringValue(current.element_ref)
  );
  return current;
}

function resolveActionSelector(params) {
  const direct = stringValue(params.selector);
  if (direct) {
    return direct;
  }
  const fromRef = selectorFromElementRef(stringValue(params.element_ref));
  if (fromRef) {
    return fromRef;
  }
  const fromHint = selectorFromElementHint(objectValue(params.element_hint));
  if (fromHint) {
    return fromHint;
  }
  return selectorFromElementResolver(objectValue(params.element_resolver));
}

function selectorFromElementHint(hint) {
  const selector = stringValue(hint.selector);
  if (selector) {
    return selector;
  }
  return selectorFromLocatorDescriptor(hint);
}

function selectorFromElementResolver(resolver) {
  const selector = stringValue(resolver.selector);
  if (selector) {
    return selector;
  }
  const elementRef = stringValue(resolver.element_ref);
  if (elementRef) {
    const refSelector = selectorFromElementRef(elementRef);
    if (refSelector) {
      return refSelector;
    }
  }
  for (const key of ['match_plan', 'locator_plan']) {
    for (const candidate of arrayValue(resolver[key])) {
      const candidateSelector = selectorFromLocatorDescriptor(objectValue(candidate));
      if (candidateSelector) {
        return candidateSelector;
      }
    }
  }
  return '';
}

function firstNonEmpty(...values) {
  for (const value of values) {
    const current = stringValue(value);
    if (current) {
      return current;
    }
  }
  return '';
}

function uploadPaths(params) {
  const paths = arrayValue(params.paths).map((item) => stringValue(item)).filter(Boolean);
  if (paths.length === 0) {
    throw httpError(400, 'paths is required');
  }
  return paths;
}

function normalizeFillFields(rawFields) {
  return arrayValue(rawFields)
    .map((field) => normalizeRequestParams(field))
    .filter((field) => Object.keys(field).length > 0);
}

function selectValues(params) {
  const values = arrayValue(params.values).map((item) => stringValue(item)).filter(Boolean);
  if (values.length > 0) {
    return values;
  }
  const value = stringValue(params.value);
  if (value) {
    return [value];
  }
  throw httpError(400, 'value or values is required');
}

function dragStartLocatorParams(params) {
  return {
    ref: stringValue(params.start_ref),
    element_ref: stringValue(params.start_ref),
    hint: normalizeRequestParams(params.start_hint),
    element_hint: normalizeRequestParams(params.start_hint),
    resolver: normalizeRequestParams(params.start_resolver),
    element_resolver: normalizeRequestParams(params.start_resolver),
    selector: stringValue(params.start_selector),
    wait_ms: intValue(params.wait_ms),
    tab_index: intValue(params.tab_index)
  };
}

function dragEndLocatorParams(params) {
  return {
    ref: stringValue(params.end_ref),
    element_ref: stringValue(params.end_ref),
    hint: normalizeRequestParams(params.end_hint),
    element_hint: normalizeRequestParams(params.end_hint),
    resolver: normalizeRequestParams(params.end_resolver),
    element_resolver: normalizeRequestParams(params.end_resolver),
    selector: stringValue(params.end_selector),
    wait_ms: intValue(params.wait_ms),
    tab_index: intValue(params.tab_index)
  };
}

function dragResolverOutcome(startOutcome, endOutcome) {
  return normalizeResolverOutcome(startOutcome) || normalizeResolverOutcome(endOutcome);
}

function popupClickWaitTimeout(timeout) {
  const fallback = 3000;
  if (timeout <= 0) {
    return fallback;
  }
  return Math.max(500, Math.min(timeout, 5000));
}

function clickActionTimeout(timeout, popupLikely) {
  if (!popupLikely) {
    return timeout;
  }
  if (timeout <= 0) {
    return 4000;
  }
  return Math.max(1000, Math.min(timeout, 4000));
}

async function popupIntentForLocator(locator) {
  try {
    const details = await locator.evaluate((element) => {
      if (!element || typeof element.getAttribute !== 'function') {
        return {};
      }
      return {
        tag: String(element.tagName || '').trim().toLowerCase(),
        target: String(element.getAttribute('target') || '').trim().toLowerCase(),
        form_target: String(element.getAttribute('formtarget') || '').trim().toLowerCase(),
        onclick: String(element.getAttribute('onclick') || '')
      };
    });
    const target = stringValue(details.target).toLowerCase();
    const formTarget = stringValue(details.form_target).toLowerCase();
    const onclick = stringValue(details.onclick);
    return {
      likely: target === '_blank' || formTarget === '_blank' || /\bwindow\.open\s*\(/i.test(onclick),
      tag: stringValue(details.tag),
      target,
      form_target: formTarget
    };
  } catch {
    return { likely: false };
  }
}

async function waitForPopupFromClick(page, timeout) {
  try {
    return await page.waitForEvent('popup', { timeout });
  } catch (err) {
    if (timeoutErrorLike(err)) {
      return null;
    }
    throw err;
  }
}

async function settlePopupAfterClick(popup, timeout) {
  if (!popup) {
    return;
  }
  try {
    await popup.waitForLoadState('domcontentloaded', { timeout });
  } catch (err) {
    if (!timeoutErrorLike(err)) {
      throw err;
    }
  }
}

async function clickLocatorWithPopupHandling(page, locator, clickOptions, popupLikely, popupTimeout) {
  const beforeURL = safePageURL(page);
  try {
    if (!popupLikely) {
      await locator.click(clickOptions);
      return withClickNavigationWaitOutcome(page, beforeURL, clickOptions, { popup: null, retried: false });
    }
    let result = await clickLocatorPopupOnce(page, locator, clickOptions, popupTimeout);
    if (result.popup || !result.retryable) {
      return withClickNavigationWaitOutcome(page, beforeURL, clickOptions, result);
    }
    await page.waitForTimeout(150);
    result = await clickLocatorPopupOnce(page, locator, clickOptions, popupTimeout);
    result.retried = true;
    if (result.popup || !result.error) {
      return withClickNavigationWaitOutcome(page, beforeURL, clickOptions, result);
    }
    throw result.error;
  } catch (err) {
    if (err && typeof err === 'object') {
      err.navigationWaitOutcome = clickNavigationWaitFailureOutcome(err, clickOptions);
    }
    throw err;
  }
}

function withClickNavigationWaitOutcome(page, beforeURL, clickOptions, result) {
  return {
    ...result,
    navigationWait: clickNavigationWaitSuccessOutcome(page, beforeURL, clickOptions, result)
  };
}

function clickNavigationWaitSuccessOutcome(page, beforeURL, clickOptions, result) {
  const afterURL = safePageURL(page);
  if (result?.popup) {
    return { status: 'passed', detail: 'popup domcontentloaded observed' };
  }
  if (clickOptions?.noWaitAfter) {
    return { status: 'skipped', detail: 'click used noWaitAfter; same-page navigation wait intentionally skipped' };
  }
  if (beforeURL && afterURL && beforeURL !== afterURL) {
    return { status: 'passed', detail: `playwright click completed; url_changed=true final_url=${afterURL}` };
  }
  return { status: 'passed', detail: 'playwright click completed; no same-page navigation observed' };
}

function clickNavigationWaitFailureOutcome(err, clickOptions = {}) {
  const message = errorMessage(err);
  if (clickOptions?.noWaitAfter) {
    return { status: 'skipped', detail: `click used noWaitAfter; navigation wait skipped after click failure: ${message}` };
  }
  if (browserActionNavigationWaitError(err)) {
    return { status: 'failed', detail: message };
  }
  return { status: 'skipped', detail: `click did not complete; navigation wait not reached: ${message}` };
}

function browserActionNavigationWaitError(err) {
  const message = errorMessage(err).toLowerCase();
  return message.includes('navigation') ||
    message.includes('waiting until') ||
    message.includes('load state') ||
    message.includes('waiting for scheduled navigations');
}

function safePageURL(page) {
  try {
    return page && typeof page.url === 'function' ? stringValue(page.url()) : '';
  } catch {
    return '';
  }
}

async function clickLocatorPopupOnce(page, locator, clickOptions, popupTimeout) {
  const initialPageCount = currentPages().length;
  const popupPromise = waitForPopupFromClick(page, popupTimeout);
  let clickError = null;
  try {
    await locator.click(clickOptions);
  } catch (err) {
    clickError = err;
  }
  const popup = await popupPromise;
  const openedPopup = await popupOrNewestManagedPage(popup, initialPageCount, popupTimeout);
  if (openedPopup) {
    return {
      popup: openedPopup,
      retried: false,
      retryable: false,
      error: null
    };
  }
  if (clickError && !timeoutErrorLike(clickError)) {
    throw clickError;
  }
  return {
    popup: null,
    retried: false,
    retryable: true,
    error: clickError
  };
}

async function popupOrNewestManagedPage(popup, initialPageCount, popupTimeout) {
  if (popup) {
    registerManagedPage(popup);
    await settlePopupAfterClick(popup, popupTimeout);
    return popup;
  }
  const pages = currentPages();
  if (pages.length <= initialPageCount) {
    return null;
  }
  const newest = pages[pages.length - 1];
  if (!newest || newest.isClosed()) {
    return null;
  }
  registerManagedPage(newest);
  await settlePopupAfterClick(newest, popupTimeout);
  return newest;
}

async function applyFillField(locator, field, timeout) {
  const fieldType = normalizeFillFieldType(firstNonEmpty(stringValue(field.type), await inferFillFieldType(locator)));
  const values = fillFieldValues(field);
  switch (fieldType) {
    case 'checkbox':
    case 'radio':
      await locator.setChecked(fillFieldCheckedValue(values[0]));
      return;
    case 'select':
    case 'select-one':
    case 'select-multiple':
      await locator.selectOption(values);
      return;
    case 'file':
      throw httpError(400, 'file inputs must use browser.upload');
    default:
      return fillLocatorTextWithFallback(locator, values[0], timeout);
  }
  return { fallbackUsed: false };
}

function timeoutErrorLike(err) {
  const message = errorMessage(err).toLowerCase();
  return message.includes('timeout') || message.includes('timed out');
}

async function fillLocatorTextWithFallback(locator, text, timeout) {
  try {
    if (timeout > 0) {
      await locator.fill(text, { timeout });
    } else {
      await locator.fill(text);
    }
    return { fallbackUsed: false };
  } catch (err) {
    if (!timeoutErrorLike(err)) {
      throw err;
    }
    const fallbackUsed = await locator.evaluate((element, value) => {
      if (!element) {
        return false;
      }
      const isInput = element instanceof HTMLInputElement;
      const isTextarea = element instanceof HTMLTextAreaElement;
      if (!isInput && !isTextarea && !element.isContentEditable) {
        return false;
      }
      try {
        if (typeof element.focus === 'function') {
          element.focus();
        }
      } catch {}
      if (isInput || isTextarea) {
        element.value = value;
      } else {
        element.textContent = value;
      }
      element.dispatchEvent(new Event('input', { bubbles: true }));
      element.dispatchEvent(new Event('change', { bubbles: true }));
      return true;
    }, text).catch(() => false);
    if (!fallbackUsed) {
      throw err;
    }
    return { fallbackUsed: true };
  }
}

async function clearLocatorEditableValue(locator) {
  return locator.evaluate((element) => {
    if (!element) {
      return false;
    }
    const isInput = element instanceof HTMLInputElement;
    const isTextarea = element instanceof HTMLTextAreaElement;
    if (!isInput && !isTextarea && !element.isContentEditable) {
      return false;
    }
    try {
      if (typeof element.focus === 'function') {
        element.focus();
      }
    } catch {}
    if (isInput || isTextarea) {
      element.value = '';
    } else {
      element.textContent = '';
    }
    return true;
  }).catch(() => false);
}

async function typeLocatorTextWithFallback(locator, text, timeout) {
  await clearLocatorEditableValue(locator);
  const keyboardTimeout = timeout > 0 ? Math.min(timeout, 4000) : 4000;
  try {
    if (keyboardTimeout > 0) {
      await locator.type(text, { delay: 20, timeout: keyboardTimeout });
    } else {
      await locator.type(text, { delay: 20 });
    }
    return { fallbackUsed: false, mode: 'keyboard_type' };
  } catch (err) {
    if (!timeoutErrorLike(err)) {
      throw err;
    }
    const fillResult = await fillLocatorTextWithFallback(locator, text, timeout);
    return { fallbackUsed: fillResult.fallbackUsed, mode: fillResult.fallbackUsed ? 'dom_value_fallback' : 'fill' };
  }
}

async function submitLocatorWithFallback(locator, timeout) {
  const submitTimeout = timeout > 0 ? Math.min(timeout, 5000) : 5000;
  try {
    await locator.press('Enter', { timeout: submitTimeout, noWaitAfter: true });
    return { fallbackUsed: false };
  } catch (err) {
    if (!timeoutErrorLike(err)) {
      throw err;
    }
    const fallbackUsed = await locator.evaluate((element) => {
      if (!element) {
        return false;
      }
      try {
        if (typeof element.focus === 'function') {
          element.focus();
        }
      } catch {}
      const init = {
        key: 'Enter',
        code: 'Enter',
        keyCode: 13,
        which: 13,
        bubbles: true,
        cancelable: true
      };
      try {
        element.dispatchEvent(new KeyboardEvent('keydown', init));
        element.dispatchEvent(new KeyboardEvent('keypress', init));
        element.dispatchEvent(new KeyboardEvent('keyup', init));
      } catch {}
      const form = typeof element.closest === 'function' ? element.closest('form') : null;
      if (form) {
        try {
          if (typeof form.requestSubmit === 'function') {
            form.requestSubmit();
          } else if (typeof form.submit === 'function') {
            form.submit();
          }
        } catch {}
      }
      return true;
    }).catch(() => false);
    if (!fallbackUsed) {
      throw err;
    }
    return { fallbackUsed: true };
  }
}

function fillFieldValues(field) {
  const values = arrayValue(field.values).map((item) => stringValue(item)).filter(Boolean);
  if (values.length > 0) {
    return values;
  }
  const value = stringValue(field.value);
  if (value) {
    return [value];
  }
  throw httpError(400, 'fill field value or values is required');
}

function fillFieldCheckedValue(value) {
  const normalized = stringValue(value).trim().toLowerCase();
  if (!normalized) {
    return false;
  }
  return !['0', 'false', 'off', 'no', 'unchecked'].includes(normalized);
}

function normalizeFillFieldType(value) {
  const normalized = stringValue(value).trim().toLowerCase();
  if (normalized === 'dropdown') {
    return 'select';
  }
  return normalized;
}

async function inferFillFieldType(locator) {
  return await locator.evaluate((element) => {
    const tag = String(element?.tagName || '').toLowerCase();
    const type = String(element?.getAttribute?.('type') || '').toLowerCase();
    if (tag === 'select') {
      return element.multiple ? 'select-multiple' : 'select-one';
    }
    if (type) {
      return type;
    }
    return tag;
  }).catch(() => '');
}

function selectorFromLocatorDescriptor(descriptor) {
  const selector = stringValue(descriptor.selector);
  if (selector) {
    return selector;
  }
  const nativeRef = stringValue(descriptor.native_ref || descriptor.element_ref);
  if (nativeRef) {
    const refSelector = selectorFromElementRef(nativeRef);
    if (refSelector) {
      return refSelector;
    }
  }
  const href = stringValue(descriptor.href);
  if (href) {
    return `a[href=${JSON.stringify(href)}]`;
  }
  const placeholder = stringValue(descriptor.placeholder);
  if (placeholder) {
    return `[placeholder=${JSON.stringify(placeholder)}]`;
  }
  const tag = stringValue(descriptor.tag);
  const type = stringValue(descriptor.type);
  if (tag && type) {
    return `${tag}[type=${JSON.stringify(type)}]`;
  }
  if (tag) {
    return tag;
  }
  if (type) {
    return `[type=${JSON.stringify(type)}]`;
  }
  return '';
}

function selectorFromElementRef(ref) {
  const payload = decodeElementRef(ref);
  if (!payload) {
    return '';
  }
  return selectorFromLocatorDescriptor(payload);
}

function decodeElementRef(ref) {
  const value = stringValue(ref);
  if (!value) {
    return null;
  }
  try {
    if (value.startsWith(metaElementRefPrefix)) {
      const decoded = decodeBase64URL(value.slice(metaElementRefPrefix.length));
      return normalizeRequestParams(JSON.parse(decoded.toString('utf8')));
    }
    if (value.startsWith(cssElementRefPrefix)) {
      return { selector: decodeBase64URL(value.slice(cssElementRefPrefix.length)).toString('utf8').trim() };
    }
  } catch {
    return null;
  }
  return null;
}

function decodeBase64URL(value) {
  let normalized = stringValue(value).replace(/-/g, '+').replace(/_/g, '/');
  while (normalized.length % 4 !== 0) {
    normalized += '=';
  }
  return Buffer.from(normalized, 'base64');
}

function snapshotFrameSelectorExistsEvaluate(payload) {
  const selector = String(payload?.requestedSelector || '').trim();
  if (!selector) {
    return false;
  }
  return Boolean(resolveSnapshotRoot(selector));

  function resolveSnapshotRoot(currentSelector) {
    const direct = safeQuerySelector(document, currentSelector);
    if (direct) {
      return direct;
    }
    return querySelectorDeep(document.body, currentSelector);
  }

  function querySelectorDeep(rootNode, currentSelector) {
    const queue = [rootNode];
    while (queue.length > 0) {
      const current = queue.shift();
      if (!current) {
        continue;
      }
      const found = safeQuerySelector(current, currentSelector);
      if (found) {
        return found;
      }
      for (const child of Array.from(current.children || [])) {
        if (child.shadowRoot) {
          queue.push(child.shadowRoot);
        }
        queue.push(child);
      }
    }
    return null;
  }

  function safeQuerySelector(rootNode, currentSelector) {
    if (!rootNode || typeof rootNode.querySelector !== 'function') {
      return null;
    }
    try {
      return rootNode.querySelector(currentSelector);
    } catch {
      return null;
    }
  }
}

function snapshotFrameEvaluate(payload) {
  const requestedSelector = String(payload?.requestedSelector || '').trim();
  const limitChars = Number(payload?.limitChars || 0);
  const limitElements = Number(payload?.limitElements || 0) > 0 ? Number(payload.limitElements) : 32;
  const root = requestedSelector ? resolveSnapshotRoot(requestedSelector) : document.body;
  if (!root) {
    return {
      snapshot: '',
      elements: [],
      note: `selector not found: ${requestedSelector || 'body'}`
    };
  }
  const text = String(root.innerText || root.textContent || '').trim();
  const selectorCounts = new Map();
  const elements = collectInteractiveElements(root, limitElements).map((el, index) => {
    const resolvedSelector = elementSelector(el);
    const selectorIndex = selectorCounts.get(resolvedSelector) || 0;
    selectorCounts.set(resolvedSelector, selectorIndex + 1);
    return {
      Index: index + 1,
      Role: elementRole(el),
      Tag: el.tagName.toLowerCase(),
      Label: elementLabel(el).slice(0, 120),
      Selector: resolvedSelector,
      selector_index: selectorIndex,
      Type: el.getAttribute('type') || '',
      Href: el.getAttribute('href') || '',
      Placeholder: el.getAttribute('placeholder') || ''
    };
  });
  return {
    snapshot: limitChars > 0 ? text.slice(0, limitChars) : text,
    truncated: limitChars > 0 && text.length > limitChars,
    elements
  };
  function elementSelector(el) {
    if (el.id) {
      return `#${el.id}`;
    }
    const name = el.getAttribute('name');
    if (name) {
      return `${el.tagName.toLowerCase()}[name="${name}"]`;
    }
    return el.tagName.toLowerCase();
  }

  function elementRole(el) {
    const explicit = String(el.getAttribute('role') || '').trim().toLowerCase();
    if (explicit) {
      return explicit;
    }
    const tag = String(el.tagName || '').trim().toLowerCase();
    switch (tag) {
      case 'a':
        return el.getAttribute('href') ? 'link' : '';
      case 'button':
        return 'button';
      case 'textarea':
        return 'textbox';
      case 'select':
        return 'combobox';
      case 'summary':
        return 'button';
      case 'input': {
        const type = String(el.getAttribute('type') || '').trim().toLowerCase();
        switch (type) {
          case 'button':
          case 'submit':
          case 'reset':
          case 'image':
            return 'button';
          case 'checkbox':
            return 'checkbox';
          case 'radio':
            return 'radio';
          case 'range':
            return 'slider';
          default:
            return '';
        }
      }
      default:
        return '';
    }
  }

  function elementLabel(el) {
    const ariaLabelledBy = String(el.getAttribute('aria-labelledby') || '').trim();
    if (ariaLabelledBy) {
      const pieces = [];
      ariaLabelledBy.split(/\s+/).forEach((id) => {
        const node = document.getElementById(id);
        const text = String(node && (node.textContent || node.innerText || '') || '').trim();
        if (text) {
          pieces.push(text);
        }
      });
      if (pieces.length > 0) {
        return pieces.join(' ');
      }
    }
    const ariaLabel = String(el.getAttribute('aria-label') || '').trim();
    if (ariaLabel) {
      return ariaLabel;
    }
    if (Array.isArray(el.labels) || (el.labels && typeof el.labels.length === 'number')) {
      const pieces = [];
      for (const labelEl of Array.from(el.labels || [])) {
        const text = String(labelEl && (labelEl.textContent || labelEl.innerText || '') || '').trim();
        if (text) {
          pieces.push(text);
        }
      }
      if (pieces.length > 0) {
        return pieces.join(' ');
      }
    }
    const parentLabel = typeof el.closest === 'function' ? el.closest('label') : null;
    const parentLabelText = String(parentLabel && (parentLabel.textContent || parentLabel.innerText || '') || '').trim();
    if (parentLabelText) {
      return parentLabelText;
    }
    return String(
      el.getAttribute('placeholder') ||
      el.getAttribute('name') ||
      el.getAttribute('title') ||
      el.textContent ||
      ''
    ).trim();
  }

  function resolveSnapshotRoot(currentSelector) {
    const direct = safeQuerySelector(document, currentSelector);
    if (direct) {
      return direct;
    }
    return querySelectorDeep(document.body, currentSelector);
  }

  function collectInteractiveElements(rootNode, limit) {
    const out = [];
    visitNode(rootNode);
    return out;

    function visitNode(node) {
      if (!node || out.length >= limit) {
        return;
      }
      if (node.nodeType === Node.DOCUMENT_FRAGMENT_NODE || node.nodeType === Node.DOCUMENT_NODE) {
        for (const child of Array.from(node.children || [])) {
          visitNode(child);
          if (out.length >= limit) {
            return;
          }
        }
        return;
      }
      if (node.nodeType !== Node.ELEMENT_NODE) {
        return;
      }
      const element = node;
      if (element.matches?.('a,button,input,textarea,select,[role="button"]')) {
        out.push(element);
        if (out.length >= limit) {
          return;
        }
      }
      if (element.shadowRoot) {
        visitNode(element.shadowRoot);
        if (out.length >= limit) {
          return;
        }
      }
      for (const child of Array.from(element.children || [])) {
        visitNode(child);
        if (out.length >= limit) {
          return;
        }
      }
    }
  }

  function querySelectorDeep(rootNode, currentSelector) {
    const queue = [rootNode];
    while (queue.length > 0) {
      const current = queue.shift();
      if (!current) {
        continue;
      }
      const found = safeQuerySelector(current, currentSelector);
      if (found) {
        return found;
      }
      for (const child of Array.from(current.children || [])) {
        if (child.shadowRoot) {
          queue.push(child.shadowRoot);
        }
        queue.push(child);
      }
    }
    return null;
  }

  function safeQuerySelector(rootNode, currentSelector) {
    if (!rootNode || typeof rootNode.querySelector !== 'function') {
      return null;
    }
    try {
      return rootNode.querySelector(currentSelector);
    } catch {
      return null;
    }
  }
}

function stringValue(value) {
  if (typeof value !== 'string') {
    return '';
  }
  return value.trim();
}

function safeCall(fn) {
  try {
    return fn();
  } catch {
    return '';
  }
}

async function safeAsyncString(fn) {
  try {
    const value = await fn();
    return stringValue(value);
  } catch {
    return '';
  }
}

function intValue(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return 0;
  }
  return Math.trunc(num);
}

function envString(name) {
  return stringValue(process.env[name]);
}

function envInt(name) {
  return intValue(process.env[name]);
}

function splitCSV(raw) {
  if (!raw) {
    return [];
  }
  return raw.split(',').map((item) => item.trim()).filter(Boolean);
}

async function readJSON(req) {
  const chunks = [];
  for await (const chunk of req) {
    chunks.push(chunk);
  }
  const raw = Buffer.concat(chunks).toString('utf8').trim();
  if (!raw) {
    return {};
  }
  try {
    return JSON.parse(raw);
  } catch {
    throw httpError(400, 'invalid json body');
  }
}

function sendResult(res, result) {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify({ result }));
}

function sendError(res, statusCode, message, resolverOutcome = null) {
  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json');
  const payload = { error: message };
  if (resolverOutcome) {
    payload.resolver_outcome = resolverOutcome;
  }
  res.end(JSON.stringify(payload));
}

function httpError(statusCode, message) {
  const err = new Error(message);
  err.statusCode = statusCode;
  return err;
}

function errorStatus(err) {
  if (err && Number.isInteger(err.statusCode) && err.statusCode >= 400) {
    return err.statusCode;
  }
  return 500;
}

function errorMessage(err) {
  if (!err) {
    return 'unknown error';
  }
  if (typeof err.message === 'string' && err.message.trim() !== '') {
    return err.message.trim();
  }
  return String(err);
}

function urlOrigin(value) {
  try {
    return new URL(stringValue(value)).origin;
  } catch {
    return '';
  }
}

function urlPath(value) {
  try {
    const pathname = new URL(stringValue(value)).pathname;
    return pathname || '/';
  } catch {
    return '';
  }
}

function comparableURL(value) {
  const current = stringValue(value);
  if (!current) {
    return '';
  }
  try {
    const parsed = new URL(current);
    const pathname = parsed.pathname || '/';
    return stringValue(parsed.origin + pathname + parsed.search);
  } catch {
    return current;
  }
}

function timestampTag() {
  return new Date().toISOString().replace(/[:.]/g, '').replace('T', '-').replace('Z', '');
}
