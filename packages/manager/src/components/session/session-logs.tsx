import React from 'react';
import { ISessionLog, SessionLogType } from '@polo/common/state/models';
import dayjs from 'dayjs';
import './session-logs.scss';
import { observer } from 'mobx-react-lite';
import { useScroll } from './scroll-hook';
import { FixedSizeList as List } from 'react-window';
import { Button } from '@polo/common/components/elements/button/button';
import { DownArrowIcon } from '@polo/common/components/elements/icons/down-arrow/down-arrow-icon';
import classnames from 'classnames';
const { parse } = require('ansicolor');

const colorsByLogType: {
    [key in SessionLogType]: string;
} = {
    [SessionLogType.TRACE]: "#d48ead",
    [SessionLogType.DEBUG]: "#88c0d0",
    [SessionLogType.INFO]: "#5e81ac",
    [SessionLogType.WARN]: "#ebcb8b",
    [SessionLogType.ERROR]: "#bf616a",
    [SessionLogType.CRITICAL]: "#ad1c2b",
    [SessionLogType.STDIN]: "#AAA",
    [SessionLogType.STDOUT]: "#a3be8c",
    [SessionLogType.STDERR]: "#d08770"
}

const SessionLogsRow =  ({ index, style, data }: { index: number; style: React.CSSProperties; data: ISessionLog[] }) => {
    const log = data[index];
    const color = colorsByLogType[log.type];
    return <p style={style} className="mx-2 lg:mx-0 leading-relaxed text-sm whitespace-nowrap max-w-full min-w-0 flex items-center" key={index}>
        <span className="hidden lg:inline-block uppercase text-xs font-mono min-w-24 px-3 flex-shrink-0 select-none">[{dayjs(log.when).format('HH:mm:ss')}]</span>
        <span className="hidden lg:inline-block uppercase text-xs w-16 min-w-16 flex-shrink-0 select-none" style={{ color }}>{log.type}:</span>
        {parseMessage(log.message)}
    </p>
}

export const SessionLogs = observer((props: {
    logs: ISessionLog[];
    onLogsProportionChanged: (proportions: number) => void;
    failed?: boolean;
}) => {
    const itemsHeight = 22;
    const {
        contentRef, containerRef, listRef,
        onScroll, onMouseEnter, onMouseMove, onMouseLeave,
        contentHeight,
        downArrowVisible, onDownArrowClick
    } = useScroll(props.onLogsProportionChanged, itemsHeight, props.logs ? props.logs.length : 0);

    if (!props.logs) return null;

    return <div ref={containerRef}
        className={classnames(
            'lg:m-2 lg:mt-5 flex-grow mt-10 mb-10 lg:mb-36 min-w-0 min-h-0 overflow-auto relative session-logs',
            props.failed && 'border-r-4 border-l-4 border-nord11'
        )}
        onMouseEnter={onMouseEnter}
        onMouseMove={onMouseMove}
        onMouseLeave={onMouseLeave}>
        <Button
            ghost
            bgVisible
            absolute
            hidden={!downArrowVisible}
            className="right-6 bottom-4 z-50"
            onClick={onDownArrowClick}
            icon={<DownArrowIcon className="h-8 w-8" />} />
        <List
            ref={listRef}
            outerRef={contentRef}
            className="text-nord-3 dark:text-nord4 tracking-wide __container --auto-hide"
            height={contentHeight}
            itemCount={props.logs.length}
            itemSize={itemsHeight}
            itemData={props.logs}
            width="100%"
            overscanCount={15}
            onScroll={onScroll}>
            {SessionLogsRow}
        </List>
    </div>
})


function parseMessage(message: string) {
    try {
        const parsed = parse(message);
        return <span className="whitespace-nowrap">{parsed.spans
            .map(({ css, text }: { css: string; text: string; }, i: number) => {
                const styles = css.split(';').reduce<{ [k: string]: string }>((acc, style) => {
                    if (!style.trim())
                        return acc;
                    const [key, value] = style.trim().split(':');
                    acc[key] = value;
                    return acc;
                }, {});
                return <span style={{ ...styles, paddingRight: '2px' }} key={i}>{parseSpaces(text)}</span>
            })}</span>;
    } catch {
        return <span>{parseSpaces(message)}</span>
    }
}

function parseSpaces(message: string) {
    const acc: JSX.Element[] = [];
    let spaces = 0;
    let chars = '';
    for (let i = 0; i < message.length; i++) {
        const char = message[i];
        if (char === ' ') {
            if (chars != '') {
                acc.push(<span key={acc.length}>{chars}</span>)
            }
            chars = '';
            spaces++;
        } else if (char === '\t') {
            if (chars != '') {
                acc.push(<span key={acc.length}>{chars}</span>)
            }
            chars = '';
            spaces += 4;
        } else {
            if (spaces > 0) {
                acc.push(<span key={acc.length} style={{ whiteSpace: 'pre' }}>{" ".repeat(spaces)}</span>)
            }
            spaces = 0;
            chars += char;
        }
    }
    if (chars != '') {
        acc.push(<span key={acc.length}>{chars}</span>)
    }
    if (spaces > 0) {
        acc.push(<span key={acc.length} style={{ whiteSpace: 'pre' }}>{" ".repeat(spaces)}</span>)
    }
    return acc;
}