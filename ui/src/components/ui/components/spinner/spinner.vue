<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div class="v-spinner">
        <svg :height="size" :width="size" xmlns="http://www.w3.org/2000/svg">
            <g fill="none" fill-rule="evenodd" :stroke-width="strokeWidth">
                <circle stroke="var(--spinner-background)" :cx="half" :cy="half" :r="innerHalf" />
                <circle
                    stroke="var(--accent)"
                    stroke-linecap="round"
                    :cx="half"
                    :cy="half"
                    :r="innerHalf"
                    :stroke-dasharray="dash"
                />
            </g>
        </svg>
    </div>
</template>

<script lang="ts">
import {computed} from "vue";

export default {
    props: {
        size: {
            default: 48,
            type: Number,
        },
        strokeWidth: {
            default: 3,
            type: Number,
        },
    },
    setup(props: Record<string, any>): Record<string, unknown> {
        const dash = computed(() => {
            const q = (Math.PI * half.value) / 2;
            return `${q},${q * 3}`;
        });

        const half = computed(() => props.size / 2);

        const innerHalf = computed(() => (props.size - props.strokeWidth) / 2);

        return {
            dash,
            half,
            innerHalf,
        };
    },
};
</script>

<style lang="scss" scoped>
@keyframes rotate {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(359deg);
    }
}

.v-spinner {
    align-items: center;
    display: inline-flex;
    justify-content: center;

    svg {
        animation: rotate 0.65s infinite linear;
    }
}
</style>
