import { APIPayload, APIRequestResult } from "@polo/common/api/common";
import {getLogsWSURL} from '@polo/common/api/session';
import { ISession, ISessionLog } from "@polo/common/state/models";
import { values } from "mobx";
import { useRef, useEffect } from "react";
import { useHistory } from "react-router-dom";
import {ITheme, Terminal} from 'xterm';
import {AttachAddon} from 'xterm-addon-attach';
import {FitAddon} from 'xterm-addon-fit';
import {SerializeAddon} from 'xterm-addon-serialize';
import {Unicode11Addon} from 'xterm-addon-unicode11';
import {WebLinksAddon} from 'xterm-addon-web-links';

export const useSessionRetrieval = (
    retrieveFailedSession: (uuid: string) => Promise<APIPayload<ISession>>,
    onSessionFail: () => void,
    session: ISession,
) => {
    const interval = useRef<NodeJS.Timeout | null>(null);
    const history = useHistory();

    useEffect(() => {

        const sessionStatusRetrieval = async () => {

            const logs: ISessionLog[] = values(session.logs) as any;

            let lastLogUUID: string | undefined = undefined;

            if (logs.length) {
                lastLogUUID = logs[logs.length - 1].uuid;
            }

            let fetchFailed = false;
            try {
                const logsRequest = await session.retrieveLogsAndStatus(lastLogUUID);
                if (logsRequest.result === APIRequestResult.SUCCEEDED) {
                    interval.current = setTimeout(() => sessionStatusRetrieval(), 1000);
                } else {
                    fetchFailed = true;
                }
            } catch (e) {
                console.error(e);
                fetchFailed = true;
            }

            let redirectToDashboard = false;
            if (fetchFailed) {
                redirectToDashboard = true;
                try {
                    const failedSessionRequest = await retrieveFailedSession(session.uuid);
                    if (failedSessionRequest.result === APIRequestResult.SUCCEEDED) {
                        redirectToDashboard = false;
                        onSessionFail();
                    }
                } catch (e) {
                    console.error(e);
                }
            }

            if (redirectToDashboard)
                history.push(`/_polo_/`);
        };

        sessionStatusRetrieval();

        return () => {
            if (interval.current)
                clearTimeout(interval.current);
        }
    }, [])
}

export const useSessionTerminalRetrieval = (
    session: ISession,
    container: React.MutableRefObject<HTMLDivElement>,
    retrieveFailedSession: (uuid: string) => Promise<APIPayload<ISession>>,
    onSessionFail: () => void,
) => {
    const interval = useRef<NodeJS.Timeout | null>(null);
    const history = useHistory();

    useEffect(() => {
        if (!container.current)
            return;

        const sessionStatusRetrieval = async () => {

            let fetchFailed = false;
            try {
                // Refreshes the status, which will be checked automatically
                // from the component (or one of its ancestors) containing this hook
                await session.retrieveStatus();
                await session.retrieveIntegrationsStatus();
                interval.current = setTimeout(() => sessionStatusRetrieval(), 1000);
            } catch (e) {
                console.error(e);
                fetchFailed = true;
            }

            let redirectToDashboard = false;
            if (fetchFailed) {
                redirectToDashboard = true;
                try {
                    const failedSessionRequest = await retrieveFailedSession(session.uuid);
                    if (failedSessionRequest.result === APIRequestResult.SUCCEEDED) {
                        redirectToDashboard = false;
                        onSessionFail();
                    }
                } catch (e) {
                    console.error(e);
                }
            }

            if (redirectToDashboard)
                history.push(`/_polo_/`);
        };

        sessionStatusRetrieval();

        const darkMode = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;

        let theme: ITheme = {
            background: 'rgba(255, 255, 255, 0)',
            foreground: 'rgb(107, 114, 128)',
            selectionBackground: '#D8DEE9',
            selectionForeground: '#2E3440',
            brightBlue: '#5E81AC',
            brightCyan: '#D8DEE9',
            brightGreen: '#7B8B6F',
            brightMagenta: '#B48EAD',
            brightYellow: '#CDAE7A',
            brightRed: '#BF616A',
            brightBlack: 'rgba(46, 52, 64, .3)',
        };

        if (darkMode) {
            theme = {
                background: 'rgba(255, 255, 255, 0)',
                foreground: 'rgb(229, 231, 235)',
                blue: '#81A1C1',
                black: '#D8DEE9',
                cyan: '#5E81AC',
                brightBlue: '#5E81AC',
                brightCyan: '#5E81AC',
                brightGreen: '#F4F8E8',
                brightMagenta: '#B48EAD',
                brightYellow: '#EBCB8B',
                brightRed: '#BF616A',
                selectionBackground: '#2E3440',
                selectionForeground: '#D8DEE9',
                brightBlack: 'rgba(229, 233, 240, .35)'
            };
        }

        const terminal = new Terminal({
            cursorBlink: false,
            disableStdin: true,
            cols: 128,
            allowProposedApi: true,
            allowTransparency: true,
            theme,
            fontFamily: 'Source Code Pro, monospace',
            fontWeight: '400',
            fontSize: 13,
            lineHeight: 1.15,

        });
        terminal.open(container.current);

        const protocol = location.protocol === 'https:' ? 'wss://' : 'ws://';
        const url = `${protocol}${location.host}${getLogsWSURL(session.uuid)}`;
        const ws = new WebSocket(url);

        const attachAddon = new AttachAddon(ws);
        const fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.loadAddon(new SerializeAddon());
        terminal.loadAddon(new WebLinksAddon());
        terminal.loadAddon(new Unicode11Addon());

        ws.onopen = function() {
            terminal.loadAddon(attachAddon);
            terminal.focus();
            setTimeout(() => fitAddon.fit(), 0);
        };

        terminal.onResize(function(event) {
            const rows = event.rows;
            const cols = event.cols;
            const size = JSON.stringify({cols: cols, rows: rows + 1});
            const send = new TextEncoder().encode("\x01" + size);
            ws.send(send);
        });

        const onWindowResize = () => {
            fitAddon.fit();
        };

        window.addEventListener('resize', onWindowResize);

        return () => {
            window.removeEventListener('resize', onWindowResize);
            
            if (interval.current)
                clearTimeout(interval.current);
        }
    }, [container])
}