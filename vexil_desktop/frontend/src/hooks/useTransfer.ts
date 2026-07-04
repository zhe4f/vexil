import { useState, useEffect, useCallback } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type { TransferView } from "../types";

interface UseTransferReturn {
    view: TransferView | null;
    isActive: boolean;
    reset: () => void;
}

export function useTransfer(): UseTransferReturn {
    const [view, setView] = useState<TransferView | null>(null);
    const [isActive, setIsActive] = useState(false);

    useEffect(() => {
        const cancel1 = EventsOn("transfer:update", (v: TransferView) => {
            setView(v);
            setIsActive(true);
        });
        const cancel2 = EventsOn("transfer:closed", () => setIsActive(false));
        return () => { cancel1(); cancel2(); };
    }, []);

    const reset = useCallback(() => setIsActive(false), []);
    return { view, isActive, reset };
}