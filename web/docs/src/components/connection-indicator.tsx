"use client";

import { useEffect, useState } from "react";

export function ConnectionIndicator() {
    const [connected, setConnected] = useState(false);

    useEffect(() => {
        // TODO: Poll /ping endpoint to check server connectivity
        // For now, assume connected
        setConnected(true);
    }, []);

    return (
        <div
            className={`h-3 w-3 rounded-full ${connected ? "animate-pulse bg-green-500" : "bg-red-500"}`}
            title={connected ? "Server Connected" : "Server Disconnected"}
        />
    );
}
