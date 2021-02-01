import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React from 'react';

type TProps = {
    branches: string[];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ServiceBranches = observer((props: TProps) => {
    return <>
        <h4 className="mt-2 mb-1 text-sm text-gray-500 uppercase">Branches:</h4>
        {sortBranches(values(props.branches) as string[]).map((branch, key) =>
            <div
                key={key}
                className="grid grid-cols-12 items-center h-12 gap-2">
                <span className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis col-span-6" title={branch}>{branch}</span>
                <div className="col-span-2">
                    <div className="text-xs text-gray-500 uppercase">Last author</div>
                    <div className="text-sm">lorem ipsum</div>
                </div>
                <div className="col-span-2">
                    <div className="text-xs text-gray-500 uppercase">Last update</div>
                    <div className="text-sm">13-41-5740</div>
                </div>
                <span className="col-span-2 text-center">
                    <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch)}>Create session</span>
                </span>
            </div>)}
    </>;
});

function sortBranches(branches: string[]): string[] {
    const preferred = [
        'master',
        'main',
        'hotfix',
        'develop',
        'dev',
        'feature'
    ];

    let result: string[] = [];
    let length = branches.length;

    for (const pref of preferred) {
        for (let i = 0; i < length; i++) {
            const branch = branches[i];
            if (branch.toLowerCase().startsWith(pref.toLowerCase())) {
                result.push(branch);
                branches.splice(i, 1)
                length--;
                i--;
            }
        }
    }

    return result.concat(branches);
}