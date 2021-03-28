import React from 'react';
import dayjs from 'dayjs';

type TProps = {
    commitMessage: string;
    commitAuthorName: string;
    commitAuthorEmail: string;
    commitDate: string;
};
export const CommitMessage = (props: TProps) => {
    return <blockquote className="relative px-2 lg:px-4 py-3 italic text-gray-500 dark:text-gray-400 z-10 bg-nord6 shadow-md dark:bg-nord-5 leading-loose">
        <p className="text-sm pb-1 flex flex-col min-h-0 max-h-32 overflow-y-auto">
            {props.commitMessage
                .split('\n')
                .map((line, key) =>
                    <span key={key}>{line}</span>)}
        </p>
        <cite className="flex items-center">
            <span className="mb-1 text-sm font-bold italic flex-1 mr-3">~ {props.commitAuthorName}</span>
            <span className="mb-1 text-xs font-light italic flex flex-nowrap">
                <span className="sm:hidden">({dayjs(props.commitDate).format('DD MMM HH:mm')})</span>
                <span className="hidden sm:block">({props.commitAuthorEmail} - {dayjs(props.commitDate).format('DD MMM HH:mm')})</span>
            </span>
        </cite>
    </blockquote>
}