const CURRENT_LEDGER_ID_KEY = 'ezbookkeeping.currentLedgerId';

export function getCurrentLedgerId(): string {
    try {
        const value = localStorage.getItem(CURRENT_LEDGER_ID_KEY);

        if (!value || value === '0' || !/^\d+$/.test(value)) {
            return '';
        }

        return value;
    } catch {
        return '';
    }
}

// Ledger ids are 64-bit integers and exceed JavaScript's safe integer range,
// so they must be kept and sent as strings to avoid precision loss.
export function setCurrentLedgerId(ledgerId: string | number): void {
    try {
        const value = ledgerId ? ledgerId.toString() : '';

        if (!value || value === '0' || !/^\d+$/.test(value)) {
            localStorage.removeItem(CURRENT_LEDGER_ID_KEY);
        } else {
            localStorage.setItem(CURRENT_LEDGER_ID_KEY, value);
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
