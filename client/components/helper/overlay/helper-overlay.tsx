import React, { memo, useLayoutEffect } from 'react';
import './helper-overlay.scss';
import { HelperStatus } from '../contexts';

const setStyleToSiblings = (setStyle: (style: CSSStyleDeclaration) => void) => {
    const body = document.getElementsByTagName('body')[0];
    if (body) {
        Array.from(body.children)
            .forEach((child: HTMLElement) => {
                if (!child.matches('#polo-session-helper'))
                    setStyle(child.style);
            })
    }
}

const idleStatus = HelperStatus.EXPIRED;

export const HelperOverlay = memo((props: { status: HelperStatus }) => {

    useLayoutEffect(() => {
        if (props.status === idleStatus) {
            setStyleToSiblings(style => style.filter = 'blur(5px)');
        }
    }, [props.status]);

    if (props.status !== idleStatus)
        return null;

    return <>
        <div className="helper-overlay__component">
            <span className="__text">
                The session has expired
            </span>
            <br /><br />
            <a className="__link" href="/_polo_">Return to dashboard</a>
        </div>
    </>;
}, (prev, next) => prev.status === next.status);