const CURRENT_LEDGER_ID_KEY = 'ezbookkeeping.currentLedgerId';

export function getCurrentLedgerId(): number {
    try {
        const value = localStorage.getItem(CURRENT_LEDGER_ID_KEY);

        if (!value) {
            return 0;
        }

        const ledgerId = parseInt(value, 10);
        return isNaN(ledgerId) || ledgerId <= 0 ? 0 : ledgerId;
    } catch {
        return 0;
    }
}

export function setCurrentLedgerId(ledgerId: number): void {
    try {
        if (!ledgerId || ledgerId <= 0) {
            localStorage.removeItem(CURRENT_LEDGER_ID_KEY);
        } else {
            localStorage.setItem(CURRENT_LEDGER_ID_KEY, ledgerId.toString());
        }
    } catch {
        // ignore
    }
}

export function clearCurrentLedgerId(): void {
    try {
        localStorage.removeItem(CURRENT_LEDGER_ID_KEY);
    } catch {
        // ignore
    }
}
