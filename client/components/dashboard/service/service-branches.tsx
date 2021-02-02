import { IServiceBranchModel } from '@/state/models';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React from 'react';
import dayjs from 'dayjs';

type TProps = {
    branches: IServiceBranchModel[];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ServiceBranches = observer((props: TProps) => {
    return <>
        <h4 className="mt-2 mb-1 text-sm text-gray-500 uppercase">Branches:</h4>
        {sortBranches(props.branches).map((branch, key) =>
            <div
                key={key}
                className="grid grid-cols-12 items-center h-12 gap-2">
                <span className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis col-span-3" title={branch.name}>{branch.name}</span>
                <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis col-span-3">{branch.message}</div>
                <div className="col-span-3">
                    <div className="text-xs text-gray-500 uppercase">Last author</div>
                    <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.author}</div>
                </div>
                <div className="col-span-1">
                    <div className="text-xs text-gray-500 uppercase">Updated</div>
                    <div className="text-sm whitespace-nowrap">{dayjs(branch.date).format('DD MMM HH:mm')}</div>
                </div>
                <span className="col-span-2 text-center">
                    <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                </span>
            </div>)}
    </>;
});

function sortBranches(branches: IServiceBranchModel[]): IServiceBranchModel[] {
    const preferred = [
        'master',
        'main',
        'hotfix',
        'develop',
        'dev',
        'feature'
    ];

    let result: IServiceBranchModel[] = [];
    let length = branches.length;

    for (const pref of preferred) {
        for (let i = 0; i < length; i++) {
            const branch = branches[i];
            if (branch.name.toLowerCase().startsWith(pref.toLowerCase())) {
                result.push(branch);
                branches.splice(i, 1)
                length--;
                i--;
            }
        }
    }

    return result.concat(branches);
}