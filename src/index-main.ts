import uaParser from 'ua-parser-js';

function isMobileDevice(): boolean {
    const userAgent = navigator.userAgent || '';

    if (!userAgent) {
        return false;
    }

    // Some fnOS gateway / embedded webviews report a desktop-like user agent on mobile devices,
    // so also fall back to the common mobile tokens and the actual viewport.
    if (/Mobile|Android|iPhone|iPad|iPod|HarmonyOS|Windows Phone|MicroMessenger/i.test(userAgent)) {
        return true;
    }

    const uaParseRet = uaParser(userAgent);

    if (uaParseRet && uaParseRet.device) {
        const device = uaParseRet.device;

        if (device.type === 'mobile' || device.type === 'wearable' || device.type === 'embedded') {
            return true;
        }
    }

    try {
        if (navigator.maxTouchPoints > 1 && window.matchMedia('(max-width: 1024px)').matches) {
            return true;
        }
    } catch {
        // ignore
    }

    return false;
}

function navigate(type: string): void {
    if (__EZBOOKKEEPING_IS_PRODUCTION__) {
        window.location.replace(`${type}#/`);
    } else {
        window.location.replace(`${type}.html#/`);
    }
}

if (isMobileDevice()) {
    navigate('mobile');
} else {
    navigate('desktop');
}
