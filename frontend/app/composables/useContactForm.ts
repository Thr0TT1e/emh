import { create } from '@bufbuild/protobuf';

import { isConnectError, isRateLimited, retryAfterSeconds } from '~/lib/errors';
import {
  CreateContactMessageRequestSchema,
  type CreateContactMessageRequest,
} from '~/sdk/emh/v1/contact_pb';

export interface UseContactFormOptions {
  /** Вызывается после успешной отправки (для аналитики и т.п.) */
  onSuccess?: (event: { subject: string; pageUrl: string }) => void;
}

export type ContactFormStatus = 'idle' | 'submitting' | 'success' | 'error';

export interface ContactFormErrors {
  name?: string;
  email?: string;
  message?: string;
  consent?: string;
}

export interface ContactSubjectOption {
  label: string;
  value: string;
}

const NAME_MIN = 2;
const NAME_MAX = 200;
const EMAIL_MAX = 254;
const MESSAGE_MIN = 10;
const MESSAGE_MAX = 5000;

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

export function useContactForm(options: UseContactFormOptions = {}) {
  const { contact } = useApi();

  const form = reactive({
    name: '',
    email: '',
    subject: '',
    message: '',
    consent: false,
    honeypot: '',
  });

  const errors = ref<ContactFormErrors>({});
  const status = ref<ContactFormStatus>('idle');
  const globalError = ref('');
  const retryAfter = ref(0);

  const subjectOptions: ContactSubjectOption[] = [
    { label: 'Сотрудничество', value: 'Сотрудничество' },
    { label: 'Вопрос по проекту', value: 'Вопрос по проекту' },
    { label: 'Предложение по разработке', value: 'Предложение по разработке' },
    { label: 'Другое', value: 'Другое' },
  ];

  function clearErrors(): void {
    errors.value = {};
    globalError.value = '';
    retryAfter.value = 0;
  }

  function validate(): boolean {
    const next: ContactFormErrors = {};

    const name = form.name.trim();
    const email = form.email.trim();
    const message = form.message.trim();

    if (!name) {
      next.name = 'Укажите имя.';
    } else if (name.length < NAME_MIN) {
      next.name = `Имя должно содержать минимум ${NAME_MIN} символа.`;
    } else if (name.length > NAME_MAX) {
      next.name = `Имя не должно превышать ${NAME_MAX} символов.`;
    }

    if (!email) {
      next.email = 'Укажите email для обратной связи.';
    } else if (email.length > EMAIL_MAX) {
      next.email = `Email не должен превышать ${EMAIL_MAX} символов.`;
    } else if (!EMAIL_REGEX.test(email)) {
      next.email = 'Укажите корректный email.';
    }

    if (!message) {
      next.message = 'Напишите сообщение.';
    } else if (message.length < MESSAGE_MIN) {
      next.message = `Сообщение должно содержать минимум ${MESSAGE_MIN} символов.`;
    } else if (message.length > MESSAGE_MAX) {
      next.message = `Сообщение не должно превышать ${MESSAGE_MAX} символов.`;
    }

    if (!form.consent) {
      next.consent = 'Необходимо согласие на обработку данных.';
    }

    errors.value = next;

    return Object.keys(next).length === 0;
  }

  async function submit(): Promise<void> {
    if (status.value === 'submitting') {
      return;
    }

    // Honeypot: если скрытое поле заполнено, вероятно, это бот.
    // Внешне показываем успех, но запрос не отправляем.
    if (form.honeypot.trim().length > 0) {
      status.value = 'success';
      return;
    }

    if (!validate()) {
      status.value = 'idle';
      return;
    }

    // Если rate-limit ещё активен, не отправляем
    if (retryAfter.value > 0) {
      return;
    }

    status.value = 'submitting';
    clearErrors();

    try {
      const pageUrl = import.meta.client ? window.location.href : '';

      const request: CreateContactMessageRequest = create(CreateContactMessageRequestSchema, {
        name: form.name.trim(),
        email: form.email.trim().toLowerCase(),
        subject: form.subject.trim(),
        message: form.message.trim(),
        pageUrl,
        honeypot: form.honeypot.trim(),
        consent: true,
      });

      await contact.createContactMessage(request);

      status.value = 'success';

      options.onSuccess?.({
        subject: form.subject.trim() || 'Без темы',
        pageUrl,
      });

      form.name = '';
      form.email = '';
      form.subject = '';
      form.message = '';
      form.consent = false;
      form.honeypot = '';
    } catch (err) {
      status.value = 'error';

      if (isRateLimited(err)) {
        const seconds = retryAfterSeconds(err);
        retryAfter.value = seconds;
        globalError.value = `Слишком много сообщений. Попробуйте через ${seconds} сек.`;

        // Запускаем таймер обратного отсчёта
        if (import.meta.client) {
          const interval = setInterval(() => {
            retryAfter.value = Math.max(0, retryAfter.value - 1);
            if (retryAfter.value <= 0) {
              clearInterval(interval);
              globalError.value = '';
            }
          }, 1000);
        }
      } else if (isConnectError(err)) {
        // Обработка validation errors (invalid_argument)
        if (err.code === 3) {
          // Code.InvalidArgument
          globalError.value = err.rawMessage || 'Проверьте поля формы.';
        } else {
          globalError.value =
            'Не удалось отправить сообщение. Попробуйте позже или напишите на info@neverforgotten.ru.';
        }
      } else {
        globalError.value =
          'Не удалось отправить сообщение. Попробуйте позже или напишите на info@neverforgotten.ru.';
      }
    }
  }

  function reset(): void {
    form.name = '';
    form.email = '';
    form.subject = '';
    form.message = '';
    form.consent = false;
    form.honeypot = '';

    clearErrors();
    status.value = 'idle';
  }

  return {
    form,
    errors,
    status,
    globalError,
    retryAfter,
    subjectOptions,
    submit,
    reset,
  };
}
